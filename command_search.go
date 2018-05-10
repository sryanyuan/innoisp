package main

import (
	"fmt"
	"os"
	"spf13/cobra"
	"time"
)

type searchStatistic struct {
	pageSearched      int
	indexPageSearched int
	searchTimes       int
	startTm           int64
}

type searchOptions struct {
	file   string
	key    int
	pksize int
}

func newSearchCommand() *cobra.Command {
	var options searchOptions
	c := &cobra.Command{
		Use:   "search",
		Short: "search the primary key in innodb ibd file",
		Long:  "search the primary key in innodb ibd file",
		Run: func(cmd *cobra.Command, args []string) {
			doSearch(cmd, &options)
		},
	}

	c.Flags().StringVarP(&options.file, "file", "f", "", "innodb table space file path")
	c.Flags().IntVarP(&options.key, "key", "k", -1, "which key to search")
	c.Flags().IntVarP(&options.pksize, "pksize", "p", 8, "primary key size (BIGINT=8,INT=4,SINT=2,TINT=1)")

	return c
}

func doSearch(cmd *cobra.Command, options *searchOptions) {
	if "" == options.file {
		fmt.Println("No input file specified")
		return
	}
	if options.key < 0 {
		fmt.Println("Invalid search key")
		return
	}
	if options.pksize <= 0 {
		fmt.Println("Invalid primary key size")
		return
	}

	f, err := os.Open(options.file)
	if nil != err {
		fmt.Println("Open file error ", err)
		return
	}
	defer f.Close()

	searchKey(f, options)
}

func searchKey(f *os.File, options *searchOptions) {
	// Get the file size
	fst, err := f.Stat()
	if nil != err {
		fmt.Printf("Get file info error %v\r\n", err)
		return
	}
	pageCount := int(fst.Size()) / (16 * 1024)
	if pageCount < 4 {
		// No index page
		fmt.Printf("No index page found\r\n")
		return
	}
	fmt.Printf("File %s has %d page(s)\r\n", f.Name(), pageCount)
	fmt.Println("Searching for file segment inode page ...")

	inodePage, err := readPageFromFile(f, 2, &parsePageOptions{})
	if nil != err {
		fmt.Printf("Read file segment inode page data error %v\r\n", err)
		return
	}
	fmt.Printf("File segment inode page found at index %d\r\n", inodePage.no)
	// Locate to root index page from inode page
	// every index occupy one inode as internal (non-leaf) node
	fmt.Println("Searching for root index page ...")
	rootIndexInode := inodePage.inode.inodes[0]
	if nil == rootIndexInode {
		fmt.Println("Can't find root index inode")
		return
	}
	// Check used
	if rootIndexInode.magicNumber != inodeEntryMagicNumber {
		fmt.Println("Inode not initialized")
		return
	}
	if 0xffffffff == rootIndexInode.fragmentArrayEntry[0] {
		fmt.Println("Index root page not allocated")
		return
	}
	if int(rootIndexInode.fragmentArrayEntry[0]) >= pageCount {
		fmt.Println("Root index page index out of range")
		return
	}
	inodeUsedCnt := 0
	for _, v := range rootIndexInode.fragmentArrayEntry {
		if v != 0xffffffff {
			inodeUsedCnt++
		}
	}
	fmt.Printf("Root index page found at index %d, %d inode used\r\n",
		int(rootIndexInode.fragmentArrayEntry[0]), inodeUsedCnt)
	// Load the root index page
	fmt.Printf("Loading root index page at index %d\r\n",
		rootIndexInode.fragmentArrayEntry[0])
	rootIndexPage, err := readPageFromFile(f, int(rootIndexInode.fragmentArrayEntry[0]), &parsePageOptions{
		parseRecords: true,
		pksize:       options.pksize,
	})
	if nil != err {
		fmt.Printf("Read root index page from file error %v\r\n", err)
		return
	}

	// Search for indexes
	var searchSt searchStatistic
	searchSt.startTm = time.Now().UnixNano() / 1e6
	searchIndexes(f, rootIndexPage, options, &searchSt)
}

func searchIndexes(f *os.File, page *Page, options *searchOptions, st *searchStatistic) {
	fmt.Printf("Search directory slots of page %d level %d, directory slots count %d\r\n",
		page.no, page.pheader.level, len(page.dslots))
	st.pageSearched++
	if page.pheader.level != 0 {
		st.indexPageSearched++
	}
	// Search page slots, the first is infimum (min than all records),
	// and the last if supremum (max than all records)
	// Get the value array to do binary search
	// non-leaf page, including root index page and internal page
	// Just search with range
	var slot *DSlots
	if 2 == len(page.dslots) {
		// Just read the supremum slot own records
		slot = page.dslots[1]
	} else {
		// Search the slots to find the right slot (binary search)
		slot = searchDslots(page.dslots[:], options.key, st)
	}
	if 1 == slot.owned && slot.rctype == recorderTypeSupremum {
		// Supremum slot not own any record except it self, so no record found
		fmt.Printf("Record not found\r\n")
		return
	}
	// Search key in the slot owned recorders
	var rc *compactRecorder
	if page.pheader.level == 0 {
		// Leaf node, the recorder is the row data
		rc = searchSlotEqual(slot, options.key, st)
	} else {
		// Non-leaf node, find the next page
		rc = searchSlotRange(slot, options.key, st)
	}
	if nil == rc {
		// Not found in nonleaf index page
		fmt.Printf("Record not found\r\n")
		return
	}
	// Read the pointer page and search again
	if page.pheader.level == 0 {
		// We already found the recorder
		fmt.Printf("Recorder found, page <%d> header offset <0x%04X> data offset<0x%04X>\r\n",
			page.no, rc.offset, rc.fieldDataOffset)
		fmt.Printf("Statistics: Page searched <%d> index page searched <%d> search times <%d> cost <%d ms>\r\n",
			st.pageSearched, st.indexPageSearched, st.searchTimes, time.Now().UnixNano()/1e6-st.startTm)
	} else {
		// We should find the recorder in the next page
		nextPage, err := readPageFromFile(f, int(rc.pageptr), &parsePageOptions{
			parseRecords: true,
			pksize:       options.pksize,
		})
		if nil != err {
			fmt.Printf("Read next page from file error %v\r\n", err)
			return
		}
		searchIndexes(f, nextPage, options, st)
	}
}

// In range: [,)
func recordInRangeLeftClosedRightOpen(key int, lrc *compactRecorder, rrc *compactRecorder) bool {
	k64 := int64(key)

	if lrc.header.recordType != recorderTypeInfimum {
		// Infimum system recorder is less than all recorder
		if k64 < lrc.key {
			return false
		}
	}
	if rrc.header.recordType == recorderTypeSupremum {
		return true
	}
	if k64 < rrc.key {
		return true
	}
	return false
}

// In range: (,]
func recordInRangeLeftOpenRightClosed(key int, lrc *compactRecorder, rrc *compactRecorder) bool {
	k64 := int64(key)

	if lrc.header.recordType != recorderTypeInfimum {
		// Infimum system recorder is less than all recorder
		if k64 <= lrc.key {
			return false
		}
	}
	if rrc.header.recordType == recorderTypeSupremum {
		return true
	}
	if k64 <= rrc.key {
		return true
	}
	return false
}

// In slot referenced node pointer recorders, the range is left closed and right open
func searchSlotRange(slot *DSlots, key int, st *searchStatistic) *compactRecorder {
	rc := slot.rcbptr
	for {
		st.searchTimes++
		if nil == rc || nil == rc.next {
			break
		}
		if recordInRangeLeftClosedRightOpen(key, rc, rc.next) {
			return rc
		}
		rc = rc.next
	}
	return nil
}

func searchSlotEqual(slot *DSlots, key int, st *searchStatistic) *compactRecorder {
	rc := slot.rcbptr
	for {
		st.searchTimes++
		if nil == rc {
			break
		}
		if rc.key == int64(key) {
			return rc
		}
		rc = rc.next
	}
	return nil
}

const (
	keyInRange = iota
	keyLessEqual
	keyGreater
)

func dslotCompare(ls *DSlots, rs *DSlots, key int) int {
	k64 := int64(key)

	if ls.rctype != recorderTypeInfimum {
		if k64 <= ls.rceptr.key {
			return keyLessEqual
		}
	}
	if rs.rctype != recorderTypeSupremum {
		if k64 > rs.rceptr.key {
			return keyGreater
		}
	}
	return keyInRange
}

func searchDslots(dslots []*DSlots, key int, st *searchStatistic) *DSlots {
	if len(dslots) == 2 {
		// Infimum and supremum, only supremum system recorder can hold
		// recorders
		return dslots[1]
	}
	starti := 0
	endi := len(dslots) - 1
	for {
		st.searchTimes++
		midi := (starti + endi + 1) / 2
		mslot := dslots[midi]
		// Is middle slot in range
		pslot := dslots[midi-1]
		cmpResult := dslotCompare(pslot, mslot, key)

		if cmpResult == keyInRange {
			return mslot
		} else if cmpResult == keyLessEqual {
			// Adjust end index
			endi = midi
		} else if cmpResult == keyGreater {
			// Adjust start index
			starti = midi
		}
	}
}
