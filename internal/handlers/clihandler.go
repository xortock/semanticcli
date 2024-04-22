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

func (handler *CliHandler) Handle(context *cli.Context) error {
	var bucketName = context.String(flags.BUCKET)
	var fileName = context.String(flags.FILE)

	// Create private s3 bucket if it does not exist yet
	var createBucketError = handler.s3Service.CreateBucketIfNotExists(&bucketName)
	if createBucketError != nil {
		return cli.Exit(createBucketError, 1)
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

	if handler.ContainsDistinctFlags(context, []string{flags.MAJOR, flags.MINOR, flags.PATCH, flags.BUILD}) {
		// handle version update request
		var message, exitCode = handler.HandleVersionUpdateRequest(context, &bucketName, &fileName)
		return cli.Exit(message, exitCode)

	} else if handler.ContainsDistinctFlags(context, []string{flags.DETAILS}) {
		var message, exitCode = handler.HandleVersionGetRequest(context, &bucketName, &fileName)
		return cli.Exit(message, exitCode)
	}

	return cli.Exit("this combination of flags is not supported", 0)
}

func (handler *CliHandler) HandleVersionUpdateRequest(context *cli.Context, bucketName *string, fileName *string) (string, int) {
	var version models.Version

	// Get the current version file and unmarshal it into a version model
	var content, getVersionError = handler.s3Service.GetFileContents(bucketName, fileName)
	if getVersionError != nil {
		return getVersionError.Error(), 1
	}
	json.Unmarshal(content, &version)

	var previousVersion = version
	// Apply all version mutations
	var flagError = handler.ApplyAllVersionFlags(context, &version)
	if flagError != nil {
		return flagError.Error(), 1
	}

	handler.ApplyVersionReset(&version, &previousVersion)

	// marshal current version and write to version file
	var newVersionContent, _ = json.Marshal(version)
	var newVersionContentError = handler.s3Service.WriteFileContents(bucketName, fileName, newVersionContent)
	if newVersionContentError != nil {
		return newVersionContentError.Error(), 1
	}

	return version.ToString(), 0
}

func (handler *CliHandler) HandleVersionGetRequest(context *cli.Context, bucketName *string, fileName *string) (string, int) {
	var version models.Version

	// Get the current version file and unmarshal it into a version model
	var content, getVersionError = handler.s3Service.GetFileContents(bucketName, fileName)
	if getVersionError != nil {
		return getVersionError.Error(), 1
	}
	json.Unmarshal(content, &version)

	return version.ToString(), 0
}

func (handler *CliHandler) ContainsDistinctFlags(context *cli.Context, validFlags []string) bool {

	validFlags = append(validFlags, flags.BUCKET)
	validFlags = append(validFlags, flags.FILE)

	var cliFlags = context.FlagNames()

	slices.Sort(validFlags)
	slices.Sort(cliFlags)

	return slices.Equal(validFlags, cliFlags)
}

func (handler *CliHandler) ApplyAllVersionFlags(context *cli.Context, version *models.Version) error {
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

func (handler *CliHandler) ApplyVersionFlag(flag string, version *int) error {
	// if the input is a - apply not change to the version
	if flag == "-" {
		return nil
	}

	// check if input is numerical and does not contains a - or + operator to indicate a change in value
	if helpers.IsNumerical(flag) && !strings.Contains(flag, "+") && !strings.Contains(flag, "-") {
		var parsedResult, firstConvertError = strconv.Atoi(flag)
		if firstConvertError != nil {
			return firstConvertError
		}
		*version = parsedResult
		return nil
	}

	// check the first character of the input is it is not a + operator it is a invalid operator
	var flagCharacters = []rune(flag)
	if flagCharacters[0] != '+' {
		return errors.New("invalid input: " + string(flagCharacters[0]) + " is not a valid operator (valid operators [ + ])")
	}

	// check all character after the first character if is not a numeric value
	var flagValues = string(flagCharacters[1:])
	if !helpers.IsNumerical(flagValues) {
		return errors.New("invalid input: " + flagValues + " is not a numeric value")
	}

	// convert the validated input to a int
	var parsedResult, firstConvertError = strconv.Atoi(flagValues)
	if firstConvertError != nil {
		return firstConvertError
	}

	*version += parsedResult
	return nil
}

func (handler *CliHandler) ApplyVersionReset(currentVersion *models.Version, previousVersion *models.Version) {
	// Reset lower version based on semantic version increase
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
