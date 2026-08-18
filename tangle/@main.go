package main

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

func main() {
	chunks := map[string]chunk{}
	<<collect>>
	<<readroot>>
	<<placed>>
	<<write>>
}

<<others>>
