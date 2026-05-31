package main

import storypack "namelesswatch/internal/storypack"

type ScaffoldOptions = storypack.ScaffoldOptions
type ValidationReport = storypack.ValidationReport

var scaffoldFiles = storypack.ScaffoldFilePaths()
var requiredPackFiles = storypack.RequiredPackFiles()

func ScaffoldPack(root string, opts ScaffoldOptions) ([]string, error) {
	return storypack.ScaffoldPack(root, opts)
}

func ValidatePack(root string) (ValidationReport, error) {
	return storypack.ValidatePack(root)
}
