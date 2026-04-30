// Command blp2webp converts BLP2 textures to WebP images.
//
// Usage:
//
//	go run ./scripts/blp2webp <dir>           # convert all .blp files in dir, write .webp alongside
//	go run ./scripts/blp2webp -o out/ <dir>   # write .webp files into out/
//	go run ./scripts/blp2webp file.blp        # convert a single file
package main

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"

	"github.com/HugoSmits86/nativewebp"
)

func main() {
	args := os.Args[1:]
	var outDir string

	// Parse -o flag
	for i := 0; i < len(args); i++ {
		if args[i] == "-o" && i+1 < len(args) {
			outDir = args[i+1]
			args = append(args[:i], args[i+2:]...)
			break
		}
	}

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "usage: blp2webp [-o outdir] <file.blp | directory>...\n")
		os.Exit(1)
	}

	var files []string
	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if info.IsDir() {
			entries, err := os.ReadDir(arg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error reading dir: %v\n", err)
				os.Exit(1)
			}
			for _, e := range entries {
				if !e.IsDir() && strings.EqualFold(filepath.Ext(e.Name()), ".blp") {
					files = append(files, filepath.Join(arg, e.Name()))
				}
			}
		} else {
			files = append(files, arg)
		}
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "no .blp files found")
		os.Exit(1)
	}

	if outDir != "" {
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "error creating output dir: %v\n", err)
			os.Exit(1)
		}
	}

	var converted, failed int
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "SKIP %s: %v\n", path, err)
			failed++
			continue
		}

		img, err := decodeBLP2(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "SKIP %s: %v\n", path, err)
			failed++
			continue
		}

		base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		dest := base + ".webp"
		if outDir != "" {
			dest = filepath.Join(outDir, dest)
		} else {
			dest = filepath.Join(filepath.Dir(path), dest)
		}

		f, err := os.Create(dest)
		if err != nil {
			fmt.Fprintf(os.Stderr, "SKIP %s: %v\n", path, err)
			failed++
			continue
		}
		if err := nativewebp.Encode(f, img, nil); err != nil {
			f.Close()
			os.Remove(dest)
			fmt.Fprintf(os.Stderr, "SKIP %s: encode error: %v\n", path, err)
			failed++
			continue
		}
		f.Close()
		fmt.Printf("OK %s → %s\n", filepath.Base(path), dest)
		converted++
	}

	fmt.Printf("\n%d converted, %d failed\n", converted, failed)
	if failed > 0 {
		os.Exit(1)
	}
}

// ---------------------------------------------------------------------------
// BLP2 decoder (copy from scripts/dbcdata/cli/blp2.go — kept standalone so
// this tool has zero internal dependencies)
// ---------------------------------------------------------------------------

func decodeBLP2(data []byte) (image.Image, error) {
	if len(data) < 148 {
		return nil, fmt.Errorf("blp2: file too small")
	}
	if string(data[0:4]) != "BLP2" {
		return nil, fmt.Errorf("blp2: bad magic %q", string(data[0:4]))
	}

	compression := data[8]
	alphaBits := data[9]
	width := int(binary.LittleEndian.Uint32(data[12:16]))
	height := int(binary.LittleEndian.Uint32(data[16:20]))

	mipOffset := binary.LittleEndian.Uint32(data[20:24])
	mipLength := binary.LittleEndian.Uint32(data[84:88])

	if mipOffset == 0 || mipLength == 0 {
		return nil, fmt.Errorf("blp2: no mipmap data")
	}
	if int(mipOffset+mipLength) > len(data) {
		return nil, fmt.Errorf("blp2: mipmap data out of bounds")
	}
	mipData := data[mipOffset : mipOffset+mipLength]

	switch compression {
	case 1:
		return decodeBLP2Paletted(data, mipData, width, height, alphaBits)
	case 2:
		switch alphaBits {
		case 0, 1:
			return decodeDXT1(mipData, width, height)
		case 8:
			return decodeDXT5(mipData, width, height)
		default:
			return nil, fmt.Errorf("blp2: unsupported DXTC alpha bits %d", alphaBits)
		}
	default:
		return nil, fmt.Errorf("blp2: unsupported compression %d", compression)
	}
}

func decodeBLP2Paletted(fileData, mipData []byte, width, height int, alphaBits uint8) (image.Image, error) {
	const paletteOffset = 148
	const paletteSize = 256 * 4
	if len(fileData) < paletteOffset+paletteSize {
		return nil, fmt.Errorf("blp2: file too small for palette")
	}

	var palette [256]color.NRGBA
	for i := range 256 {
		off := paletteOffset + i*4
		palette[i] = color.NRGBA{B: fileData[off], G: fileData[off+1], R: fileData[off+2], A: 255}
	}

	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	pixelCount := width * height
	if len(mipData) < pixelCount {
		return nil, fmt.Errorf("blp2: mipmap data too small for palette indices")
	}

	for i := range pixelCount {
		x, y := i%width, i/width
		c := palette[mipData[i]]
		switch alphaBits {
		case 0:
			c.A = 255
		case 1:
			if off := pixelCount + i/8; off < len(mipData) {
				if (mipData[off]>>(uint(i)%8))&1 == 0 {
					c.A = 0
				}
			}
		case 4:
			if off := pixelCount + i/2; off < len(mipData) {
				if i%2 == 0 {
					c.A = (mipData[off] & 0x0F) * 17
				} else {
					c.A = (mipData[off] >> 4) * 17
				}
			}
		case 8:
			if off := pixelCount + i; off < len(mipData) {
				c.A = mipData[off]
			}
		}
		img.SetNRGBA(x, y, c)
	}
	return img, nil
}

func decodeDXT1(data []byte, width, height int) (image.Image, error) {
	bx, by := (width+3)/4, (height+3)/4
	if len(data) < bx*by*8 {
		return nil, fmt.Errorf("blp2: DXT1 data too small")
	}
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	off := 0
	for by_ := range by {
		for bx_ := range bx {
			c0 := binary.LittleEndian.Uint16(data[off : off+2])
			c1 := binary.LittleEndian.Uint16(data[off+2 : off+4])
			bits := binary.LittleEndian.Uint32(data[off+4 : off+8])
			off += 8
			colors := dxt1Colors(c0, c1)
			for py := range 4 {
				for px := range 4 {
					x, y := bx_*4+px, by_*4+py
					if x >= width || y >= height {
						bits >>= 2
						continue
					}
					img.SetNRGBA(x, y, colors[bits&3])
					bits >>= 2
				}
			}
		}
	}
	return img, nil
}

func decodeDXT5(data []byte, width, height int) (image.Image, error) {
	bx, by := (width+3)/4, (height+3)/4
	if len(data) < bx*by*16 {
		return nil, fmt.Errorf("blp2: DXT5 data too small")
	}
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	off := 0
	for by_ := range by {
		for bx_ := range bx {
			a0, a1 := data[off], data[off+1]
			alphaBits := uint64(data[off+2]) | uint64(data[off+3])<<8 |
				uint64(data[off+4])<<16 | uint64(data[off+5])<<24 |
				uint64(data[off+6])<<32 | uint64(data[off+7])<<40
			alphas := dxt5Alphas(a0, a1)
			c0 := binary.LittleEndian.Uint16(data[off+8 : off+10])
			c1 := binary.LittleEndian.Uint16(data[off+10 : off+12])
			colorBits := binary.LittleEndian.Uint32(data[off+12 : off+16])
			off += 16
			colors := dxt1Colors(c0, c1)
			for py := range 4 {
				for px := range 4 {
					x, y := bx_*4+px, by_*4+py
					if x >= width || y >= height {
						colorBits >>= 2
						alphaBits >>= 3
						continue
					}
					c := colors[colorBits&3]
					colorBits >>= 2
					c.A = alphas[alphaBits&7]
					alphaBits >>= 3
					img.SetNRGBA(x, y, c)
				}
			}
		}
	}
	return img, nil
}

func dxt1Colors(c0, c1 uint16) [4]color.NRGBA {
	r0, g0, b0 := unpackRGB565(c0)
	r1, g1, b1 := unpackRGB565(c1)
	var colors [4]color.NRGBA
	colors[0] = color.NRGBA{R: r0, G: g0, B: b0, A: 255}
	colors[1] = color.NRGBA{R: r1, G: g1, B: b1, A: 255}
	if c0 > c1 {
		colors[2] = color.NRGBA{
			R: uint8((2*uint16(r0) + uint16(r1)) / 3),
			G: uint8((2*uint16(g0) + uint16(g1)) / 3),
			B: uint8((2*uint16(b0) + uint16(b1)) / 3), A: 255,
		}
		colors[3] = color.NRGBA{
			R: uint8((uint16(r0) + 2*uint16(r1)) / 3),
			G: uint8((uint16(g0) + 2*uint16(g1)) / 3),
			B: uint8((uint16(b0) + 2*uint16(b1)) / 3), A: 255,
		}
	} else {
		colors[2] = color.NRGBA{
			R: uint8((uint16(r0) + uint16(r1)) / 2),
			G: uint8((uint16(g0) + uint16(g1)) / 2),
			B: uint8((uint16(b0) + uint16(b1)) / 2), A: 255,
		}
		colors[3] = color.NRGBA{A: 0}
	}
	return colors
}

func unpackRGB565(c uint16) (r, g, b uint8) {
	r = uint8(((c >> 11) & 0x1F) * 255 / 31)
	g = uint8(((c >> 5) & 0x3F) * 255 / 63)
	b = uint8((c & 0x1F) * 255 / 31)
	return
}

func dxt5Alphas(a0, a1 uint8) [8]uint8 {
	var a [8]uint8
	a[0], a[1] = a0, a1
	if a0 > a1 {
		a[2] = uint8((6*uint16(a0) + 1*uint16(a1)) / 7)
		a[3] = uint8((5*uint16(a0) + 2*uint16(a1)) / 7)
		a[4] = uint8((4*uint16(a0) + 3*uint16(a1)) / 7)
		a[5] = uint8((3*uint16(a0) + 4*uint16(a1)) / 7)
		a[6] = uint8((2*uint16(a0) + 5*uint16(a1)) / 7)
		a[7] = uint8((1*uint16(a0) + 6*uint16(a1)) / 7)
	} else {
		a[2] = uint8((4*uint16(a0) + 1*uint16(a1)) / 5)
		a[3] = uint8((3*uint16(a0) + 2*uint16(a1)) / 5)
		a[4] = uint8((2*uint16(a0) + 3*uint16(a1)) / 5)
		a[5] = uint8((1*uint16(a0) + 4*uint16(a1)) / 5)
		a[6] = 0
		a[7] = 255
	}
	return a
}
