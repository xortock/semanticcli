package handlers

import (
	"encoding/json"
	"errors"
	"slices"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"
	"github.com/xortock/semanticloq/internal/flags"
	"github.com/xortock/semanticloq/internal/helpers"
	"github.com/xortock/semanticloq/internal/models"
	"github.com/xortock/semanticloq/internal/services"
)

type ICliHandler interface {
}

type CliHandler struct {
	s3Service services.IS3Service
}

func NewCliHandler() *CliHandler {
	return &CliHandler{
		s3Service: services.NewS3Service(),
	}
}

func (handler CliHandler) Handle(context *cli.Context) error {
	var bucketName = context.String(flags.BUCKET)
	var fileName = context.String(flags.FILE)

	// Create private s3 bucket if it does not exist yet
	var createBucketError = handler.s3Service.CreateBucketIfNotExists(&bucketName)
	if createBucketError != nil {
		cli.Exit(createBucketError, 1)
	}

	// Check if file inside that bucket with that name already exists
	// if not create a default version file of 0.0.0.0
	var doesFileExist = handler.s3Service.DoesFileExists(&bucketName, &fileName)
	if !doesFileExist {
		var defaultVersion = models.Version{}
		var defaultContent, _ = json.Marshal(defaultVersion)
		var writeDefaultContentError = handler.s3Service.WriteFileContents(&bucketName, &fileName, defaultContent)
		if writeDefaultContentError != nil {
			return cli.Exit(writeDefaultContentError, 1)
		}
	}

	var version models.Version

	// Get the current version file and unmarshal it into a version model
	var content, getVersionError = handler.s3Service.GetFileContents(&bucketName, &fileName)
	if getVersionError != nil {
		return cli.Exit(getVersionError, 1)
	}
	json.Unmarshal(content, &version)

	if handler.ContainsDistinctFlags(context, []string{flags.MAJOR, flags.MINOR, flags.PATCH, flags.BUILD}) {

		var previousVersion = version
		// Apply all version mutations
		var flagError = handler.ApplyAllVersionFlags(context, &version)
		if flagError != nil {
			return cli.Exit(flagError, 1)
		}

		handler.ApplyVersionReset(&version, &previousVersion)

		// marshal current version and write to version file
		var newVersionContent, _ = json.Marshal(version)
		var newVersionContentError = handler.s3Service.WriteFileContents(&bucketName, &fileName, newVersionContent)
		if newVersionContentError != nil {
			return cli.Exit(newVersionContentError, 1)
		}

		return cli.Exit(version.ToString(), 0)

	} else if handler.ContainsDistinctFlags(context, []string{flags.DETAILS}) {
		return cli.Exit(version.ToString(), 0)
	}

	return cli.Exit("this combination of flags is not supported", 0)
}

func (handler CliHandler) ContainsDistinctFlags(context *cli.Context, validFlags []string) bool {

	validFlags = append(validFlags, flags.BUCKET)
	validFlags = append(validFlags, flags.FILE)

	var cliFlags = context.FlagNames()

	slices.Sort(validFlags)
	slices.Sort(cliFlags)

	return slices.Equal(validFlags, cliFlags)
}

func (handler CliHandler) ApplyAllVersionFlags(context *cli.Context, version *models.Version) error {
	// Apply all version mutations
	var majorFlagError = handler.ApplyVersionFlag(context.String(flags.MAJOR), &version.Major)
	if majorFlagError != nil {
		return majorFlagError
	}

	var minorFlagError = handler.ApplyVersionFlag(context.String(flags.MINOR), &version.Minor)
	if minorFlagError != nil {
		return minorFlagError
	}

	var patchFlagError = handler.ApplyVersionFlag(context.String(flags.PATCH), &version.Patch)
	if patchFlagError != nil {
		return patchFlagError
	}

	var buildFlagError = handler.ApplyVersionFlag(context.String(flags.BUILD), &version.Build)
	if buildFlagError != nil {
		return buildFlagError
	}

	return nil
}

func (handler CliHandler) ApplyVersionFlag(flag string, version *int) error {
	if flag == "-" {
		return nil
	}

	if helpers.IsNumerical(flag) && !strings.Contains(flag, "+") && !strings.Contains(flag, "-") {
		var parsedResult, firstConvertError = strconv.Atoi(flag)
		if firstConvertError != nil {
			return firstConvertError
		}
		*version = parsedResult
		return nil
	}

	var flagCharacters = []rune(flag)
	if flagCharacters[0] != '+' {
		return errors.New("invalid operator used")
	}

	var flagValues = string(flagCharacters[1:])
	if !helpers.IsNumerical(flagValues) {
		return errors.New("Argument for: " + flag + " is to complicated")
	}

	var parsedResult, firstConvertError = strconv.Atoi(flagValues)
	if firstConvertError != nil {
		return firstConvertError
	}

	*version += parsedResult
	return nil
}

func (handler CliHandler) ApplyVersionReset(currentVersion *models.Version, previousVersion *models.Version) {
	if currentVersion.Patch > previousVersion.Patch {
		currentVersion.Build = 0
	}

	if currentVersion.Minor > previousVersion.Minor {
		currentVersion.Patch = 0
		currentVersion.Build = 0

	}

	if currentVersion.Major > previousVersion.Major {
		currentVersion.Minor = 0
		currentVersion.Patch = 0
		currentVersion.Build = 0
	}
}
