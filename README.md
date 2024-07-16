This is a program to help ease the annoyance of needing to repeatedly
go into the browser devtools to copy the auth token to reuse in Postman.

The simple idea is - on first launch, you are instructed to enter
several values that will then be saved in a CSV file in your user folder.

On subsequent runs, these values will be passed to AWS' cognito golang
package to get the IDToken, which can then be used to auth you in Postman.

The CSV file is saved under your home directory, under "~/.swing-jwt-helper.csv"
Flags

--help: This screen.

--setup: Initiate the setup screen.

--copy: Copies the token value to your clipboard in addition to printing
it in the terminal menu.
