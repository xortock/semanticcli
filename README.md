# semanticcli

Semantic loq is a cli tool meant for the distributed storage of the semantic versions of you applications. It is well suited for the CI/CD processes where application get their version post build.

# Installation

# Usage

## Set static version

```sh
semanticcli --bucket semantic-versioning --file api-version --major 1 --minor 1 --patch 0 --build 1

```
output: 
> 1.1.0.1

## Auto increase build version (also applies to other versions)

```sh
semanticcli --bucket semantic-versioning --file api-version --major 1 --minor 1 --patch 0 --build +1

```
1st output: 
> 1.1.0.2

2nd output: 
> 1.1.0.3

## Get current version

```sh
semanticcli --bucket semantic-versioning --file api-version --details

```
output: 
> 1.1.0.3

# Storage

for the storage of the version file you have the option to pass in a bucket name and file name. This will create a s3 bucket and version file accourding to the configured s3 credentials. 

# s3 Credentials

semanticcli will load configuration from environment variables, AWS shared configuration file (.aws/config), and AWS shared credentials file (.aws/credentials). To determine the Aws_Access_Key_Id, Aws_Secret_Access_Key_Id and Aws_Region
