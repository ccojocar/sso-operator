package proxy

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildNameResource(t *testing.T) {
	// 4 chars resource name
	resourceName := "name"
	// 6 chars suffix
	suffix := "suffix"

	expectedName := "name-suffix"

	name := buildName(resourceName, suffix)

	assert.Equal(t, 11, len(name))
	assert.Equal(t, expectedName, name)
}

func TestBuildNameResourceNameMoreThan63Chars(t *testing.T) {
	// 65 chars resource name
	resourceName := "aaaaaaaaaa-bbbbbbbbbb-cccccccccc-dddddddddd-eeeeeeeeee-ffffffffff"
	// 6 chars suffix
	suffix := "suffix"

	expectedName := "aaaaaaaaaa-bbbbbbbbbb-cccccccccc-dddddddddd-eeeeeeeeee-f-suffix"

	name := buildName(resourceName, suffix)

	assert.Equal(t, 63, len(name))
	assert.Equal(t, expectedName, name)
}

func TestBuildNameSuffixMoreThan63Chars(t *testing.T) {
	// 4 chars name
	resourceName := "name"
	// 65 chars resource name
	suffix := "aaaaaaaaaa-bbbbbbbbbb-cccccccccc-dddddddddd-eeeeeeeeee-ffffffffff"

	expectedName := "name-aaaaaaaaaa-bbbbbbbbbb-cccccccccc-dddddddddd-eeeeeeeeee-fff"

	name := buildName(resourceName, suffix)

	assert.Equal(t, 63, len(name))
	assert.Equal(t, expectedName, name)
}

func TestBuildNameResourceNameAndSuffixMoreThan63Chars(t *testing.T) {
	// 65 chars resource name
	resourceName := "aaaaaaaaaa-bbbbbbbbbb-cccccccccc-dddddddddd-eeeeeeeeee-ffffffffff"
	// 65 chars resource name
	suffix := "gggggggggg-hhhhhhhhhh-iiiiiiiiii-jjjjjjjjjj-kkkkkkkkkk-llllllllll"

	expectedName := "aaaaaaaaaa-bbbbbbbbbb-ccccccccc-gggggggggg-hhhhhhhhhh-iiiiiiiii"

	name := buildName(resourceName, suffix)

	assert.Equal(t, 63, len(name))
	assert.Equal(t, expectedName, name)
}

func TestBuildNameTrailingDash(t *testing.T) {
	// 65 chars resource name
	resourceName := "aaaaaaaaaa-bbbbbbbbbb-cccccccccc-dddddddddd-eeeeeeeeee-ffffff"
	// 7 chars suffix
	suffix := "sufffix"

	expectedName := "aaaaaaaaaa-bbbbbbbbbb-cccccccccc-dddddddddd-eeeeeeeeee-sufffix"

	name := buildName(resourceName, suffix)

	assert.Equal(t, 62, len(name))
	assert.Equal(t, expectedName, name)
}
