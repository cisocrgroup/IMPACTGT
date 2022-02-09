# Impgtt
Tools to generate line snippets from IMPACT ground-truth files.

## Dependencies
 * go1.17

## Install
`go install github.com/cisocrgroup/IMPACTGT/impgtt@latest`

## Usage
General usage: `impgtt [tool] [flags]`; for help `impgtt help [tool]`.

## Tools
 * `segregs` segment regions and create region-metadata files.
 * `seglines` segment region images into lines (based on the number of
   ground-truth lines).
 * `alignes` alignes ground-truth lines with ocr lines and image snippets.
 * `srv` start a simple server that lets you inspect segmented files.
 * `pack` re-packs the segmented region lines into paged directory.
