## Usage

Authenticate to Google Cloud and generate Application Default Credentials with:

```sh
gcloud auth login --update-adc
```

If required, update the value of `projectID` in `main.go` (defaults to `slo-generator-demo`).

Run the script with:

```sh
go run .
```
