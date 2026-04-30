package cli

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
)

// decodeBLP2 decodes a BLP2 texture (used by World of Warcraft) into an
// image.Image. Only the first mipmap level is decoded.
//
// Supported sub-formats:
//   - DXT1 (compression=2, alphaBits=0 or 1)
//   - DXT5 (compression=2, alphaBits=8)
//   - Paletted (compression=1)
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

	// Mipmap offsets[16] at byte 20, lengths[16] at byte 84.
	// We only use mipmap 0 (the full-size image).
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
	case 1: // Paletted
		return decodeBLP2Paletted(data, mipData, width, height, alphaBits)
	case 2: // DXTC
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

// decodeBLP2Paletted decodes a BLP2 paletted image. The palette is a 256-entry
// BGRA table at offset 148 in the file.
func decodeBLP2Paletted(fileData, mipData []byte, width, height int, alphaBits uint8) (image.Image, error) {
	const paletteOffset = 148
	const paletteSize = 256 * 4
	if len(fileData) < paletteOffset+paletteSize {
		return nil, fmt.Errorf("blp2: file too small for palette")
	}

	var palette [256]color.NRGBA
	for i := range 256 {
		off := paletteOffset + i*4
		palette[i] = color.NRGBA{
			B: fileData[off+0],
			G: fileData[off+1],
			R: fileData[off+2],
			A: 255,
		}
	}

	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	pixelCount := width * height

	// Index data comes first, then alpha data.
	if len(mipData) < pixelCount {
		return nil, fmt.Errorf("blp2: mipmap data too small for palette indices")
	}

	for i := range pixelCount {
		x := i % width
		y := i / width
		c := palette[mipData[i]]

		// Decode alpha from the data following the palette indices.
		switch alphaBits {
		case 0:
			c.A = 255
		case 1:
			alphaOff := pixelCount + i/8
			if alphaOff < len(mipData) {
				bit := (mipData[alphaOff] >> (uint(i) % 8)) & 1
				if bit == 0 {
					c.A = 0
				} else {
					c.A = 255
				}
			}
		case 4:
			alphaOff := pixelCount + i/2
			if alphaOff < len(mipData) {
				if i%2 == 0 {
					c.A = (mipData[alphaOff] & 0x0F) * 17 // scale 0-15 → 0-255
				} else {
					c.A = (mipData[alphaOff] >> 4) * 17
				}
			}
		case 8:
			alphaOff := pixelCount + i
			if alphaOff < len(mipData) {
				c.A = mipData[alphaOff]
			}
		}

		img.SetNRGBA(x, y, c)
	}

	return img, nil
}

// decodeDXT1 decodes a DXT1 (BC1) compressed texture.
// Each 4×4 block is 8 bytes: 2 reference colors + 4 bytes of 2-bit indices.
func decodeDXT1(data []byte, width, height int) (image.Image, error) {
	blocksX := (width + 3) / 4
	blocksY := (height + 3) / 4
	if len(data) < blocksX*blocksY*8 {
		return nil, fmt.Errorf("blp2: DXT1 data too small")
	}

	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	off := 0
	for by := range blocksY {
		for bx := range blocksX {
			c0 := binary.LittleEndian.Uint16(data[off : off+2])
			c1 := binary.LittleEndian.Uint16(data[off+2 : off+4])
			bits := binary.LittleEndian.Uint32(data[off+4 : off+8])
			off += 8

			colors := dxt1Colors(c0, c1)
			for py := range 4 {
				for px := range 4 {
					x := bx*4 + px
					y := by*4 + py
					if x >= width || y >= height {
						bits >>= 2
						continue
					}
					idx := bits & 3
					bits >>= 2
					img.SetNRGBA(x, y, colors[idx])
				}
			}
		}
	}
	return img, nil
}

// decodeDXT5 decodes a DXT5 (BC3) compressed texture.
// Each 4×4 block is 16 bytes: 8 bytes alpha + 8 bytes DXT1 color.
func decodeDXT5(data []byte, width, height int) (image.Image, error) {
	blocksX := (width + 3) / 4
	blocksY := (height + 3) / 4
	if len(data) < blocksX*blocksY*16 {
		return nil, fmt.Errorf("blp2: DXT5 data too small")
	}

	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	off := 0
	for by := range blocksY {
		for bx := range blocksX {
			// Alpha block (8 bytes)
			a0 := data[off]
			a1 := data[off+1]
			// 6 bytes = 48 bits of 3-bit alpha indices for 16 pixels
			alphaBits := uint64(data[off+2]) |
				uint64(data[off+3])<<8 |
				uint64(data[off+4])<<16 |
				uint64(data[off+5])<<24 |
				uint64(data[off+6])<<32 |
				uint64(data[off+7])<<40

			alphas := dxt5Alphas(a0, a1)

			// Color block (8 bytes) — same as DXT1
			c0 := binary.LittleEndian.Uint16(data[off+8 : off+10])
			c1 := binary.LittleEndian.Uint16(data[off+10 : off+12])
			colorBits := binary.LittleEndian.Uint32(data[off+12 : off+16])
			off += 16

			colors := dxt1Colors(c0, c1)
			for py := range 4 {
				for px := range 4 {
					x := bx*4 + px
					y := by*4 + py
					if x >= width || y >= height {
						colorBits >>= 2
						alphaBits >>= 3
						continue
					}

					ci := colorBits & 3
					colorBits >>= 2
					ai := alphaBits & 7
					alphaBits >>= 3

					c := colors[ci]
					c.A = alphas[ai]
					img.SetNRGBA(x, y, c)
				}
			}
		}
	}
	return img, nil
}

// dxt1Colors computes the 4-color lookup table from two 16-bit RGB565 values.
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
			B: uint8((2*uint16(b0) + uint16(b1)) / 3),
			A: 255,
		}
		colors[3] = color.NRGBA{
			R: uint8((uint16(r0) + 2*uint16(r1)) / 3),
			G: uint8((uint16(g0) + 2*uint16(g1)) / 3),
			B: uint8((uint16(b0) + 2*uint16(b1)) / 3),
			A: 255,
		}
	} else {
		colors[2] = color.NRGBA{
			R: uint8((uint16(r0) + uint16(r1)) / 2),
			G: uint8((uint16(g0) + uint16(g1)) / 2),
			B: uint8((uint16(b0) + uint16(b1)) / 2),
			A: 255,
		}
		colors[3] = color.NRGBA{A: 0} // transparent black
	}

	return colors
}

// unpackRGB565 converts a 16-bit RGB565 value to 8-bit per channel.
// Arithmetic stays in uint16 to avoid uint8 overflow on the multiply.
func unpackRGB565(c uint16) (r, g, b uint8) {
	r = uint8(((c >> 11) & 0x1F) * 255 / 31)
	g = uint8(((c >> 5) & 0x3F) * 255 / 63)
	b = uint8((c & 0x1F) * 255 / 31)
	return
}

// dxt5Alphas computes the 8-entry alpha lookup from two reference values.
func dxt5Alphas(a0, a1 uint8) [8]uint8 {
	var a [8]uint8
	a[0] = a0
	a[1] = a1
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
