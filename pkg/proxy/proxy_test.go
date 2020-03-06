package proxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildNameResource(t *testing.T) {
	// 43 chars resource name
	resourceName := "aaaaaaaaaa-bbbbbbbbbb-cccccccccc-dddddddddd"
	// 6 chars suffix
	suffix := "suffix"

	expectedName := "aaaaaaaaaa-bbbbbbbbbb-cccccccccc-dddddddddd-suffix"

	name := buildName(resourceName, suffix)

	assert.Equal(t, 50, len(name))
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
	// 65 chars suffix
	suffix := "aaaaaaaaaa-bbbbbbbbbb-cccccccccc-dddddddddd-eeeeeeeeee-ffffffffff"

	expectedName := "name-aaaaaaaaaa-bbbbbbbbbb-cccccccccc-dddddddddd-eeeeeeeeee-fff"

	name := buildName(resourceName, suffix)

	assert.Equal(t, 63, len(name))
	assert.Equal(t, expectedName, name)
}

func TestBuildNameResourceNameAndSuffixMoreThan63Chars(t *testing.T) {
	// 65 chars resource name
	resourceName := "aaaaaaaaaa-bbbbbbbbbb-cccccccccc-dddddddddd-eeeeeeeeee-ffffffffff"
	// 65 chars suffix
	suffix := "gggggggggg-hhhhhhhhhh-iiiiiiiiii-jjjjjjjjjj-kkkkkkkkkk-llllllllll"

	expectedName := "aaaaaaaaaa-bbbbbbbbbb-ccccccccc-gggggggggg-hhhhhhhhhh-iiiiiiiii"

	name := buildName(resourceName, suffix)

	assert.Equal(t, 63, len(name))
	assert.Equal(t, expectedName, name)
}

func TestBuildNameTrailingDash(t *testing.T) {
	// 61 chars resource name
	resourceName := "aaaaaaaaaa-bbbbbbbbbb-cccccccccc-dddddddddd-eeeeeeeeee-ffffff"
	// 7 chars suffix
	suffix := "sufffix"

	expectedName := "aaaaaaaaaa-bbbbbbbbbb-cccccccccc-dddddddddd-eeeeeeeeee-sufffix"

	name := buildName(resourceName, suffix)

	assert.Equal(t, 62, len(name))
	assert.Equal(t, expectedName, name)
}

func TestBuildNameNoSuffix(t *testing.T) {
	// 43 chars resource name
	resourceName := "aaaaaaaaaa-bbbbbbbbbb-cccccccccc-dddddddddd"
	// Empty string suffix
	suffix := ""

	expectedName := "aaaaaaaaaa-bbbbbbbbbb-cccccccccc-dddddddddd"

	name := buildName(resourceName, suffix)

	assert.Equal(t, 43, len(name))
	assert.Equal(t, expectedName, name)
}

func TestBuildNameNoSuffixLongResourceName(t *testing.T) {
	// 65 chars resource name
	resourceName := "aaaaaaaaaa-bbbbbbbbbb-cccccccccc-dddddddddd-eeeeeeeeee-ffffffffff"
	// Empty string suffix
	suffix := ""

	expectedName := "aaaaaaaaaa-bbbbbbbbbb-cccccccccc-dddddddddd-eeeeeeeeee-fffffff"

	name := buildName(resourceName, suffix)

	assert.Equal(t, 62, len(name))
	assert.Equal(t, expectedName, name)
}
