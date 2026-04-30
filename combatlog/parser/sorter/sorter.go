package sorter

import (
	"bufio"
	"context"
	"io"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Emyrk/chronicle/combatlog/parser/lines"
	"github.com/Emyrk/chronicle/combatlog/parser/types/combatant"
	"github.com/Emyrk/chronicle/combatlog/parser/types/realmclock"
	"github.com/Emyrk/chronicle/combatlog/parser/types/unitinfo"
	"github.com/Emyrk/chronicle/combatlog/parser/types/zone"
	"github.com/Emyrk/chronicle/combatlog/parser/wotlk"
)

type SortSummary struct {
	Earliest time.Time
	Latest   time.Time
	Total    int
	IsRaw    bool
}

type logLine struct {
	Date    time.Time
	Content string
	idx     int64
}

// SortLogs reads log lines from input, sorts them by timestamp, and writes them to output.
// When useUnixMillis is true, lines are parsed using AzerothCore's epoch-millisecond
// timestamp format (via wotlk.ExtractUnixMilli) instead of the vanilla date format.
func SortLogs(ctx context.Context, logger *slog.Logger, input io.Reader, output io.Writer, useUnixMillis bool) (SortSummary, *realmclock.Info, error) {
	sum := SortSummary{}
	buffer := make([]logLine, 0)
	var firstRealmClock *realmclock.Info

	// Use original timestamps for sorting
	liner := lines.NewLiner().WithoutTimeAdjustments()
	sc := bufio.NewScanner(input)
	c := int64(0)
	for sc.Scan() {
		if ctx.Err() != nil {
			return sum, firstRealmClock, ctx.Err()
		}

		txt := sc.Text()
		var ts time.Time
		var content string
		var err error
		if useUnixMillis {
			ts, content, err = wotlk.ExtractUnixMilli(txt)
		} else {
			ts, content, err = liner.Line(txt)
		}
		if err != nil {
			logger.Warn("skipping failed line", slog.String("line", txt), slog.String("error", err.Error()))
			continue
		}

		if isUnitLike(content) {
			sum.IsRaw = true
		}

		buffer = append(buffer, logLine{
			Date:    ts,
			Content: content,
			idx:     c,
		})
		c++

		if firstRealmClock == nil {
			if ri, err := realmclock.ParseClockInfo(content); err == nil {
				firstRealmClock = &ri
			}
		}

		if ts.Before(sum.Earliest) || sum.Earliest.IsZero() {
			sum.Earliest = ts
		}

		if ts.After(sum.Latest) {
			sum.Latest = ts
		}
		sum.Total++
	}

	// Sort primarily by timestamp. Then prioritize:
	// 1. Clock/header info lines
	// 2. Zone info lines, zone changes context for everything else
	// 3. Unit info lines, unit db should be populated asap
	// 4. Combatant lines, same idea as above
	// Finally, keep the original order for lines with identical timestamps and types
	slices.SortFunc(buffer, func(a, b logLine) int {
		am, bm := a.Date.UnixMilli(), b.Date.UnixMilli()
		if am != bm {
			return int(am - bm)
		}

		clc := compareBooleans(isClockLike(a.Content), isClockLike(b.Content))
		if clc != 0 {
			return clc
		}

		cz := compareBooleans(isZoneLike(a.Content), isZoneLike(b.Content))
		if cz != 0 {
			return cz
		}

		cu := compareBooleans(isUnitLike(a.Content), isUnitLike(b.Content))
		if cu != 0 {
			return cu
		}

		_, ac := combatant.IsCombatant(a.Content)
		_, bc := combatant.IsCombatant(b.Content)
		cc := compareBooleans(ac, bc)
		if cc != 0 {
			return cc
		}

		return int(a.idx - b.idx)
	})

	// First thing we do is insert some heading logs (vanilla format only —
	// AzerothCore timestamps are already absolute, no realm clock needed).
	if !useUnixMillis && len(buffer) > 0 && firstRealmClock != nil {
		_, _ = output.Write([]byte(
			// Knock some time off the first to guarantee it's first
			liner.FmtLine(
				buffer[0].Date.Add(time.Second*-10),
				firstRealmClock.String(),
			),
		))
		_, _ = output.Write([]byte("\n"))
	}

	for _, line := range buffer {
		if ctx.Err() != nil {
			return sum, firstRealmClock, ctx.Err()
		}

		var formatted string
		if useUnixMillis {
			formatted = strconv.FormatInt(line.Date.UnixMilli(), 10) + "  " + line.Content
		} else {
			formatted = liner.FmtLine(line.Date, line.Content)
		}
		_, err := output.Write([]byte(formatted))
		if err != nil {
			return sum, firstRealmClock, err
		}
		_, _ = output.Write([]byte("\n"))
	}

	return sum, firstRealmClock, nil
}

// isClockLike returns true for vanilla CLOCK_INFO or AzerothCore CHRONICLE_HEADER lines.
func isClockLike(s string) bool {
	_, ok := realmclock.IsClockInfo(s)
	return ok || strings.HasPrefix(s, "CHRONICLE_HEADER,")
}

// isZoneLike returns true for vanilla ZONE_INFO or AzerothCore CHRONICLE_ZONE_INFO lines.
func isZoneLike(s string) bool {
	_, ok := zone.IsZoneInfo(s)
	return ok || strings.HasPrefix(s, "CHRONICLE_ZONE_INFO,")
}

// isUnitLike returns true for vanilla UNIT_INFO or AzerothCore CHRONICLE_UNIT_INFO lines.
func isUnitLike(s string) bool {
	_, ok := unitinfo.IsUnitInfo(s)
	return ok || strings.HasPrefix(s, "CHRONICLE_UNIT_INFO,")
}

func compareBooleans(a, b bool) int {
	if a == b {
		return 0
	}
	// True should be less than false so it's sorted first
	if a && !b {
		return -1
	}
	return 1
}
