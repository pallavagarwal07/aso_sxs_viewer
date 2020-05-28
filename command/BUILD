load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library", "go_test")

go_library(
    name = "go_command_library",
    srcs = ["command.go" ],
    importpath = "github.com/googleinterns/aso_sxs_viewer/command",
    visibility = ["//visibility:private"],
)

go_binary(
    name = "command",
    embed = [":go_command_library"],
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_command_test",
    srcs = ["command_test.go"],
    embed = [":go_command_library"],
)
