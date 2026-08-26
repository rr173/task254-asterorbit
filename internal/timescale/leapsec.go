package timescale

import "time"

// leapSecondEvents 记录 TAI−UTC 跳秒生效时刻（UTC）与累计偏移量（秒）。
// 数据截至 2017-01-01（偏移 37 秒），覆盖本项目目标观测年代。
var leapSecondEvents = []struct {
	t      time.Time
	offset int
}{
	{time.Date(1972, 1, 1, 0, 0, 0, 0, time.UTC), 10},
	{time.Date(1972, 7, 1, 0, 0, 0, 0, time.UTC), 11},
	{time.Date(1973, 1, 1, 0, 0, 0, 0, time.UTC), 12},
	{time.Date(1974, 1, 1, 0, 0, 0, 0, time.UTC), 13},
	{time.Date(1975, 1, 1, 0, 0, 0, 0, time.UTC), 14},
	{time.Date(1976, 1, 1, 0, 0, 0, 0, time.UTC), 15},
	{time.Date(1977, 1, 1, 0, 0, 0, 0, time.UTC), 16},
	{time.Date(1978, 1, 1, 0, 0, 0, 0, time.UTC), 17},
	{time.Date(1979, 1, 1, 0, 0, 0, 0, time.UTC), 18},
	{time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC), 19},
	{time.Date(1981, 7, 1, 0, 0, 0, 0, time.UTC), 20},
	{time.Date(1982, 7, 1, 0, 0, 0, 0, time.UTC), 21},
	{time.Date(1983, 7, 1, 0, 0, 0, 0, time.UTC), 22},
	{time.Date(1984, 7, 1, 0, 0, 0, 0, time.UTC), 23},
	{time.Date(1985, 7, 1, 0, 0, 0, 0, time.UTC), 24},
	{time.Date(1988, 1, 1, 0, 0, 0, 0, time.UTC), 25},
	{time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC), 26},
	{time.Date(1991, 1, 1, 0, 0, 0, 0, time.UTC), 27},
	{time.Date(1992, 7, 1, 0, 0, 0, 0, time.UTC), 28},
	{time.Date(1993, 7, 1, 0, 0, 0, 0, time.UTC), 29},
	{time.Date(1994, 7, 1, 0, 0, 0, 0, time.UTC), 30},
	{time.Date(1996, 1, 1, 0, 0, 0, 0, time.UTC), 31},
	{time.Date(1997, 7, 1, 0, 0, 0, 0, time.UTC), 32},
	{time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC), 33},
	{time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC), 34},
	{time.Date(2009, 1, 1, 0, 0, 0, 0, time.UTC), 35},
	{time.Date(2012, 7, 1, 0, 0, 0, 0, time.UTC), 36},
	{time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC), 37},
}

// TAIOffsetSeconds 返回给定 UTC 时刻的 TAI−UTC 跳秒数。
func TAIOffsetSeconds(utc time.Time) int {
	off := 10
	for _, e := range leapSecondEvents {
		if !utc.Before(e.t) {
			off = e.offset
		} else {
			break
		}
	}
	return off
}
