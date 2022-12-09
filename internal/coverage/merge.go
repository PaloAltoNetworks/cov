package coverage

import (
	"fmt"
	"sort"

	"github.com/gobwas/glob"
	"golang.org/x/tools/cover"
)

func mergeProfiles(p *cover.Profile, merge *cover.Profile) error {
	if p.Mode != merge.Mode {
		return fmt.Errorf("cannot merge profiles with different modes")
	}
	// Since the blocks are sorted, we can keep track of where the last block
	// was inserted and only look at the blocks after that as targets for merge
	startIndex := 0
	var err error
	for _, b := range merge.Blocks {
		startIndex, err = mergeProfileBlock(p, b, startIndex)
		if err != nil {
			return err
		}
	}
	return nil
}

func mergeProfileBlock(p *cover.Profile, pb cover.ProfileBlock, startIndex int) (int, error) {
	sortFunc := func(i int) bool {
		pi := p.Blocks[i+startIndex]
		return pi.StartLine >= pb.StartLine && (pi.StartLine != pb.StartLine || pi.StartCol >= pb.StartCol)
	}

	i := 0
	if sortFunc(i) != true {
		i = sort.Search(len(p.Blocks)-startIndex, sortFunc)
	}
	i += startIndex
	if i < len(p.Blocks) && p.Blocks[i].StartLine == pb.StartLine && p.Blocks[i].StartCol == pb.StartCol {
		if p.Blocks[i].EndLine != pb.EndLine || p.Blocks[i].EndCol != pb.EndCol {
			return 0, fmt.Errorf("Overlap merge: %v %v %v", p.FileName, p.Blocks[i], pb)
		}
		switch p.Mode {
		case "set":
			p.Blocks[i].Count |= pb.Count
		case "count", "atomic":
			p.Blocks[i].Count += pb.Count
		default:
			return 0, fmt.Errorf("unsupported covermode: '%s'", p.Mode)
		}
	} else {
		if i > 0 {
			pa := p.Blocks[i-1]
			if pa.EndLine >= pb.EndLine && (pa.EndLine != pb.EndLine || pa.EndCol > pb.EndCol) {
				return 0, fmt.Errorf("Overlap before: %v %v %v", p.FileName, pa, pb)
			}
		}
		if i < len(p.Blocks)-1 {
			pa := p.Blocks[i+1]
			if pa.StartLine <= pb.StartLine && (pa.StartLine != pb.StartLine || pa.StartCol < pb.StartCol) {
				return 0, fmt.Errorf("Overlap after: %v %v %v", p.FileName, pa, pb)
			}
		}
		p.Blocks = append(p.Blocks, cover.ProfileBlock{})
		copy(p.Blocks[i+1:], p.Blocks[i:])
		p.Blocks[i] = pb
	}
	return i + 1, nil
}

func addProfile(profiles []*cover.Profile, p *cover.Profile) ([]*cover.Profile, error) {
	i := sort.Search(len(profiles), func(i int) bool { return profiles[i].FileName >= p.FileName })
	if i < len(profiles) && profiles[i].FileName == p.FileName {
		if err := mergeProfiles(profiles[i], p); err != nil {
			return nil, err
		}
	} else {
		profiles = append(profiles, nil)
		copy(profiles[i+1:], profiles[i:])
		profiles[i] = p
	}
	return profiles, nil
}

// MergeProfiles merge all profiles files
func MergeProfiles(ignorePatterns []string, files []string) ([]*cover.Profile, error) {

	var merged []*cover.Profile

	for _, file := range files {
		profiles, err := cover.ParseProfiles(file)
		if err != nil {
			return nil, fmt.Errorf("failed to parse profiles: %w", err)
		}
		for _, p := range profiles {
			merged, err = addProfile(merged, p)
			if err != nil {
				return nil, err
			}
		}
	}

	globs := []glob.Glob{}
	for _, pattern := range ignorePatterns {
		globs = append(globs, glob.MustCompile(pattern))
	}

	var cleanedMerged []*cover.Profile
L:
	for _, profile := range merged {
		for _, g := range globs {
			if g.Match(profile.FileName) {
				continue L
			}
		}
		cleanedMerged = append(cleanedMerged, profile)
	}

	return cleanedMerged, nil
}
