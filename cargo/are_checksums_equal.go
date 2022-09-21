package cargo

import "strings"

// AreChecksumsEqual returns true only when the given checksums match, including algorithm.
// The algorithm is given the same way that "cargo.ConfigMetadataDependency".Checksum expects,
// as "algorithm:checksum". Only exact equality is allowed, although algorithm is only checked
// if present in both of the given checksums.
func AreChecksumsEqual(c1, c2 string) bool {
	var algorithm1, algorithm2 string
	split1 := strings.Split(c1, ":")
	if len(split1) == 2 {
		algorithm1 = split1[0]
		c1 = split1[1]
	} else if len(split1) > 2 {
		return false
	}
	split2 := strings.Split(c2, ":")
	if len(split2) == 2 {
		algorithm2 = split2[0]
		c2 = split2[1]
	}
	areAlgorithmsEqual := func(a1, a2 string) bool {
		return a1 == "" || a2 == "" || a1 == a2
	}
	return areAlgorithmsEqual(algorithm1, algorithm2) && c1 == c2
}
