load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "lib.go",
        "main.go",
    ],
    importpath = "github.com/googleinterns/aso_sxs_viewer/sample",
    visibility = ["//visibility:private"],
)

go_binary(
    name = "sample",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["lib_test.go"],
    embed = [":go_default_library"],
)
