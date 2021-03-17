source = ["./dist/binary-osx.zip"]
bundle_id = "io.streamingfast.streamingfast-client.cmd"

apple_id {
  # The username when not defined is picked automatically from env var AC_USERNAME
  # The password when not defined is picked automatically from env var AC_PASSWORD
}

sign {
  application_identity = "Developer ID Application: dfuse Platform Inc. (ZG686LRL8C)"
}

notarize {
  path = "./dist/binary-osx.zip"
  bundle_id = "io.streamingfast.streamingfast-client.cmd"
}
