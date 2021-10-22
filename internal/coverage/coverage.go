package coverage

import (
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/tools/cover"
)

// Tree represents a tree of coverage profiles nodes
type Tree []node

// node
type node struct {
	Name              string  `json:"name"`
	Children          Tree    `json:"children,omitempty"`
	CoveredStatements int64   `json:"covered,omitempty"`
	TotalStaments     int64   `json:"total,omitempty"`
	Coverage          float64 `json:"coverage,omitempty"`
}

// NewTree return a new tree from cover profiles
func NewTree(profiles []*cover.Profile, files []string) Tree {

	// get all profiles we are looking for
	var tree []node

	for _, p := range profiles {
		covered, total := coverage(p)

		if len(files) > 0 {

			for _, f := range files {

				if f == "" {
					continue
				}

				if strings.Contains(p.FileName, f) {
					tree = addToTree(tree, strings.Split(p.FileName, "/"), covered, total)
				}

			}

		} else {
			tree = addToTree(tree, strings.Split(p.FileName, "/"), covered, total)
		}

	}

	// Compute all tree statements and percents
	return computeTree(tree)
}

func coverage(p *cover.Profile) (int64, int64) {
	var total, covered int64
	for _, b := range p.Blocks {
		total += int64(b.NumStmt)
		if b.Count > 0 {
			covered += int64(b.NumStmt)
		}
	}
	if total == 0 {
		return 0, 0
	}
	return covered, total
}

func addToTree(root []node, names []string, covered, total int64) []node {
	if len(names) > 0 {
		var i int
		for i = 0; i < len(root); i++ {
			if root[i].Name == names[0] { //already in tree
				break
			}
		}
		if i == len(root) {
			node := node{Name: names[0]}
			if strings.HasSuffix(names[0], ".go") {
				node.CoveredStatements = covered
				node.TotalStaments = total
			}
			root = append(root, node)
		}
		root[i].Children = addToTree(root[i].Children, names[1:], covered, total)
	}

	return root
}

func computeTree(tree []node) []node {

	for i := range tree {

		if len(tree[i].Children) > 0 {
			tree[i].Children = computeTree(tree[i].Children)
		}

		for _, c := range tree[i].Children {
			tree[i].CoveredStatements += c.CoveredStatements
			tree[i].TotalStaments += c.TotalStaments
		}

		tree[i].Coverage = float64(tree[i].CoveredStatements) / float64(tree[i].TotalStaments) * 100

	}
	return tree
}

// GetCoverage just check the first node coverage given the threshold
func (tree Tree) GetCoverage() float64 {

	if tree == nil {
		return 0
	}

	return tree[0].Coverage
}

// Fprint implement the Fpring for a node tree
func (tree Tree) Fprint(w io.Writer, root bool, padding string, threshold float64) {

	if tree == nil {
		return
	}

	if root {
		fmt.Fprintf(w, "\n")
	}

	index := 0
	for _, v := range tree {

		c := color.New()
		if threshold > 0 {
			c = color.New(color.FgGreen)
			if v.Coverage < threshold {
				c = color.New(color.FgRed)
			}
		}
		fmt.Fprintf(w, "%s%s\n", padding+getPadding(root, getBoxType(index, len(tree))), c.Sprintf("[%.0f%%] %s", v.Coverage, v.Name))
		v.Children.Fprint(w, false, padding+getPadding(root, getBoxTypeExternal(index, len(tree))), threshold)
		index++
	}
	return
}

type boxType int

const (
	regular boxType = iota
	last
	afterLast
	between
)

func (boxType boxType) String() string {
	switch boxType {
	case regular:
		return "├──"
	case last:
		return "└──"
	case afterLast:
		return "   "
	case between:
		return "│  "
	default:
		panic("invalid box type")
	}
}

func getBoxType(index int, len int) boxType {
	if index+1 == len {
		return last
	} else if index+1 > len {
		return afterLast
	}
	return regular
}

func getBoxTypeExternal(index int, len int) boxType {
	if index+1 == len {
		return afterLast
	}
	return between
}

func getPadding(root bool, boxType boxType) string {
	if root {
		return ""
	}

	return boxType.String() + " "
}
