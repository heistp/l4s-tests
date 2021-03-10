package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const emitRatio bool = true

var qdiscs = [...]string{"dualpi2",
	"pfifo(1000)",
	"pfifo(200)",
	"cnq_codel_af",
	"fq_codel",
}

type Flent struct {
	Metadata  Metadata  `json:"metadata"`
	RawValues RawValues `json:"raw_values"`
}

func (f *Flent) propValue(key string) (val string, found bool) {
	fields := strings.Fields(f.Metadata.Title)

	for _, f := range fields {
		c := strings.Index(f, ":")
		if c != -1 {
			k := f[:c]
			v := f[c+1:]
			if k == key {
				val = v
				found = true
				break
			}
		}
	}

	return
}

type Metadata struct {
	Title string `json:"TITLE"`
}

type RawValues struct {
	Upload1 []Sample `json:"TCP upload::1"`
	Upload2 []Sample `json:"TCP upload::2"`
}

type Sample struct {
	T               float64 `json:"t"`
	TCPDeliveryRate float64 `json:"tcp_delivery_rate"`
	Val             float64 `json:"val"`
}

type runInfo struct {
	bandwidth string
	qdisc     string
	cc1       string
	rtt1      string
	delivery1 float64
	cc2       string
	rtt2      string
	delivery2 float64
	jains     float64
}

func qdiscIndex(qdisc string) int {
	for i, q := range qdiscs {
		if q == qdisc {
			return i
		}
	}
	return len(qdiscs)
}

type sortRunInfo []runInfo

func (s sortRunInfo) Len() int {
	return len(s)
}

func (s sortRunInfo) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortRunInfo) Less(i, j int) bool {
	if s[i].bandwidth > s[j].bandwidth {
		return true
	}
	if s[i].bandwidth < s[j].bandwidth {
		return false
	}
	if qdiscIndex(s[i].qdisc) < qdiscIndex(s[j].qdisc) {
		return true
	}
	if qdiscIndex(s[i].qdisc) > qdiscIndex(s[j].qdisc) {
		return false
	}
	if parseRTT(s[i].rtt1) < parseRTT(s[j].rtt1) {
		return true
	}
	if parseRTT(s[i].rtt1) > parseRTT(s[j].rtt1) {
		return false
	}
	if s[i].rtt2 != "" && s[j].rtt2 != "" {
		return parseRTT(s[i].rtt2) < parseRTT(s[j].rtt2)
	}
	return false
}

func visit(files *[]string, rootDir string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			failErr(err)
		}
		if strings.HasSuffix(info.Name(), ".flent.gz") ||
			strings.HasSuffix(info.Name(), ".flent") {
			*files = append(*files, path)
		}
		return nil
	}
}

func propValue(flent *Flent, key string) (val string, found bool) {
	title := flent.Metadata.Title
	fields := strings.Fields(title)

	for _, f := range fields {
		c := strings.Index(f, ":")
		if c != -1 {
			k := f[:c]
			v := f[c+1:]
			if k == key {
				val = v
				found = true
				break
			}
		}
	}

	return
}

func jains(vals ...float64) float64 {
	var s float64
	for _, v := range vals {
		s += v
	}

	var ss float64
	for _, v := range vals {
		ss += v * v
	}

	return (s * s) / (float64(len(vals)) * ss)
}

func ssDeliveryRate(samples []Sample, ssWin int) float64 {
	last := samples[len(samples)-1]
	tlast := last.T
	tstart := tlast - float64(ssWin) - 1
	tend := tstart + float64(ssWin)

	var tdr float64
	var n int
	for _, s := range samples {
		if s.T > tstart && s.T < tend && s.TCPDeliveryRate != 0 {
			tdr += s.TCPDeliveryRate
			n++
		}
	}

	return (tdr / float64(n))
}

func process(path string, ssWin int) (info runInfo, err error) {
	var f *os.File
	if f, err = os.Open(path); err != nil {
		return
	}
	defer f.Close()

	var r io.Reader
	r = bufio.NewReader(f)

	if filepath.Ext(path) == ".gz" {
		if r, err = gzip.NewReader(bufio.NewReader(f)); err != nil {
			return
		}
	}

	d := json.NewDecoder(r)
	var flent Flent
	if err = d.Decode(&flent); err != nil {
		return
	}

	info.bandwidth, _ = flent.propValue("bandwidth")
	info.qdisc, _ = flent.propValue("qdisc")

	info.rtt1, _ = flent.propValue("rtt")
	info.rtt2, _ = flent.propValue("rtt2")

	vs, _ := flent.propValue("vs")
	ccs := strings.Split(vs, "-vs-")
	info.cc1 = ccs[0]
	info.cc2 = ccs[1]

	info.delivery1 = ssDeliveryRate(flent.RawValues.Upload1, ssWin)
	info.delivery2 = ssDeliveryRate(flent.RawValues.Upload2, ssWin)

	info.jains = jains(info.delivery1, info.delivery2)

	return
}

func parseRTT(rttstr string) (rtt int) {
	var err error
	if rtt, err = strconv.Atoi(rttstr[:len(rttstr)-2]); err != nil {
		failErr(err)
	}
	return
}

func emitHeader() {
	var lastCol string
	if emitRatio {
		lastCol = "Ratio"
	} else {
		lastCol = "Jain's"
	}

	fmt.Printf("| Rate | qdisc | CC1 (RTT) | D<sub>SS</sub>1 | CC2 (RTT) | D<sub>SS</sub>2 | %s |\n", lastCol)
	fmt.Printf("| ---- | ----- | --------- | --------------- | --------- | ----------------| -- |\n")
}

func ratio(delivery1, delivery2 float64) string {
	if delivery1 > delivery2 {
		return fmt.Sprintf("%.0f:1", delivery1/delivery2)
	}
	return fmt.Sprintf("1:%.0f", delivery2/delivery1)
}

func emitRow(r runInfo) {
	var rtt2 string
	if r.rtt2 != "" {
		rtt2 = r.rtt2
	} else {
		rtt2 = r.rtt1
	}

	var lastCol string
	if emitRatio {
		lastCol = ratio(r.delivery1, r.delivery2)
	} else {
		lastCol = fmt.Sprintf("%.3f", r.jains)
	}

	fmt.Printf("| %s | %s | %s(%s) | %.2f | %s(%s) | %.2f | %s |\n",
		r.bandwidth, r.qdisc,
		r.cc1, r.rtt1, r.delivery1,
		r.cc2, rtt2, r.delivery2,
		lastCol)
}

func failErr(err error) {
	fmt.Fprintf(os.Stderr, "error: %s\n", err)
	os.Exit(-1)
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: steady_state <results directory with .flent.gz files> <steady state trailing window, in seconds>")
		os.Exit(-1)
	}
	dir := os.Args[1]
	ssWin, err := strconv.Atoi(os.Args[2])
	if err != nil {
		failErr(err)
	}

	var files []string
	if err = filepath.Walk(dir, visit(&files, dir)); err != nil {
		failErr(err)
	}

	emitHeader()

	var runInfos []runInfo
	for _, file := range files {
		runInfo, err := process(file, ssWin)
		if err != nil {
			failErr(err)
		}
		runInfos = append(runInfos, runInfo)
	}

	sort.Sort(sortRunInfo(runInfos))

	for _, ri := range runInfos {
		emitRow(ri)
	}
}
