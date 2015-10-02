package fs

import (
	"errors"
	"fmt"

	"github.com/petar/GoLLRB/llrb"
)

// Implements a search capability in a general repository
// The search index is a node that contains a simply map of
// string (the search term) with a node -> a BTREE of terms

// The Btree node contains a serialized GoLLRB structure, with
// the word (thing to test on) being the main key
// and the values being an array of urls (paths) into the system with
// an optional version field

type SearchIndex struct {
	Node  BlockNode
	Terms map[string]BlockNode
}

type SearchTree struct {
	Node BlockNode
	Tree *llrb.LLRB
}

type SearchEntry struct {
	Term    string
	Matches []Entry
}

type Entry struct {
	Path       string
	VersionTag string
}

func (s SearchEntry) Less(than llrb.Item) bool {
	return s.Term < than.(SearchEntry).Term
}

// Ok what does the outside world see in a filesystem?

func (rfs *RootFileSystem) SearchAddTerms(area string, term []string, path string, version string) error {
	searchIndex := rfs.ChangeCache.GetSearchIndex()
	treeNode, ok := searchIndex.Terms[area]
	var searchTree *SearchTree = &SearchTree{}
	var err error
	if !ok {
		// Create a node for this tree
		treeNode = rfs.BlockHandler.GetFreeBlockNode(SEARCHTREE)
		searchIndex.Terms[area] = treeNode
		searchTree.Tree = llrb.New()
		searchTree.Node = treeNode
		// Need to save the searchTree back
		rfs.ChangeCache.SaveSearchIndex(searchIndex)
	} else {
		searchTree, err = rfs.ChangeCache.GetSearchTree(treeNode)
		if err != nil {
			return err
		}
	}
	// Now we have a search tree, add the term
	for _, t := range term {
		var searchItem = SearchEntry{t, nil}
		existingValue := searchTree.Tree.Get(searchItem)
		var pointValue SearchEntry
		if existingValue != nil {
			pointValue = existingValue.(SearchEntry)
		} else {
			pointValue = SearchEntry{t, make([]Entry, 0)}
		}
		entry := Entry{path, version}
		pointValue.Matches = append(pointValue.Matches, entry)
		searchTree.Tree.ReplaceOrInsert(pointValue)
	}
	// And save this searchTree back
	rfs.ChangeCache.SaveSearchTree(searchTree)
	return nil
}

// Add a term that can be searched on (append or create to an existing term)
func (rfs *RootFileSystem) SearchAddTerm(area string, term string, path string, version string) error {
	searchIndex := rfs.ChangeCache.GetSearchIndex()
	treeNode, ok := searchIndex.Terms[area]
	var searchTree *SearchTree = &SearchTree{}
	var err error
	if !ok {
		// Create a node for this tree
		treeNode = rfs.BlockHandler.GetFreeBlockNode(SEARCHTREE)
		searchIndex.Terms[area] = treeNode
		searchTree.Tree = llrb.New()
		searchTree.Node = treeNode
		// Need to save the searchTree back
		rfs.ChangeCache.SaveSearchIndex(searchIndex)
	} else {
		searchTree, err = rfs.ChangeCache.GetSearchTree(treeNode)
		if err != nil {
			return err
		}
	}
	// Now we have a search tree, add the term
	var searchItem = SearchEntry{term, nil}
	existingValue := searchTree.Tree.Get(searchItem)
	var pointValue SearchEntry
	if existingValue != nil {
		pointValue = existingValue.(SearchEntry)
	} else {
		pointValue = SearchEntry{term, make([]Entry, 0)}
	}
	entry := Entry{path, version}
	pointValue.Matches = append(pointValue.Matches, entry)
	searchTree.Tree.ReplaceOrInsert(pointValue)
	// And save this searchTree back
	rfs.ChangeCache.SaveSearchTree(searchTree)
	return nil
}

// Remove a term that has been added
func (rfs *RootFileSystem) SearchRemoveTerm(area string, term string, path string) {

}

// Find all entries that are between start and end for an area (set same value for ==)
func (rfs *RootFileSystem) SearchFindTerms(area string, startTerm string, endTerm string) ([]Entry, error) {
	searchIndex := rfs.ChangeCache.GetSearchIndex()
	treeNode, ok := searchIndex.Terms[area]
	if !ok {
		// No tree, so screw that
		return nil, errors.New("No search area found")
	} else {
		searchTree, err := rfs.ChangeCache.GetSearchTree(treeNode)
		if err != nil {
			return nil, err
		}
		// Now do search
		startEntry := SearchEntry{startTerm, nil}
		endEntry := SearchEntry{endTerm, nil}
		tester := make(map[Entry]struct{})
		ret := make([]Entry, 0)
		searchTree.Tree.AscendRange(startEntry, endEntry, func(item llrb.Item) bool {
			entry := item.(SearchEntry)
			fmt.Printf("Adding %v, %v", entry.Term, entry.Matches)
			for _, match := range entry.Matches {
				_, ok := tester[match]
				if !ok {
					tester[match] = struct{}{}
					ret = append(ret, match)
				}
			}
			return true
		})
		return ret, nil
	}
}
