source = ["./dist/sf-osx_darwin_amd64/sf"]
bundle_id = "io.streamingfast.streamingfast-client.cmd"

apple_id {
  # The username when not defined is picked automatically from env var AC_USERNAME
  # The password when not defined is picked automatically from env var AC_PASSWORD
}

sign {
  application_identity = "Developer ID Application: dfuse Platform Inc. (ZG686LRL8C)"
}