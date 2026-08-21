// Package report is treex's public output contract: the shape a scan takes
// when it is rendered, whether as a table for a person or as JSON for another
// program.
//
// It deliberately imports nothing from internal. A report is a plain data
// structure that a projection fills in, so the output format cannot drift with
// an internal refactor and a consumer parsing the JSON has something stable to
// rely on.
package report
