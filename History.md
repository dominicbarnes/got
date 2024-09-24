
v1.5.0 / 2024-09-24
===================

  * feat(suite): add TestCase to func signature
  * test(testdata): update test broken by empty explode change

v1.4.1 / 2024-09-24
===================

  * fix(testdata): only set explode values when not empty

v1.4.0 / 2024-09-19
===================

  * docs(readme): typo
  * lint(suite): add t.Helper()
  * docs(readme): add section for RunTestSuite
  * docs(readme): add notes about 'explode' option and 'only' skips
  * test(testdata): adding coverage to explode test for a nested directory
  * refactor(suite): use local T interface consistently

v1.3.0 / 2024-09-18
===================

  * feat(testdata): put special treatment for maps behind a struct tag option (explode) (#8)

v1.2.1 / 2024-09-16
===================

  * test(suite): fix tests on CI by using .gitkeep files in otherwise empty dirs

v1.2.0 / 2024-09-16
===================

  * feat(suite): add support for only running specific tests (the opposite of skip)

v1.1.3 / 2024-09-15
===================

  * lint: add golangci config
  * lint(thelper): use t.Helper() consistently
  * lint(tagliatelle): use consistent json tags
  * lint(perfsprint): replace fmt.Sprintf with string concat
  * fix: remove testing spew.Dump
  * lint(errorlint): properly wrap errors
  * lint(lll): maintain lines under 120 width
  * refactor: remove unnecessary context
  * refactor: remove unnecessary comments and parameters
  * lint(gocritic): convert if/else to switch
  * lint(cyclop): reduce complexity of loadDir and saveDir
  * lint(gopls): remove unnecessary type arguments
  * put lowest supported version in go.mod
  * test(suite): add coverage for RunTestSuite
  * feat(suite): add helper for common suite runner
  * test older versions of go

v1.1.2 / 2024-09-15
===================

  * docs(readme): overhaul
  * test(suite): add coverage
  * test(testdata): improve coverage
  * test(codec): add coverage

v1.1.1 / 2024-09-13
===================

  * fix broken tests
  * fix lint rules
  * add github actions to run golangci-lint
  * Update README.md

v1.1.0 / 2023-05-03
===================

  * feat(suite): add support for skipping test cases
  * test: fixing unknown-codec test case following yaml addition

v1.0.0 / 2023-04-27
===================

  * API overhaul (v1) (#7)

v0.2.0 / 2021-02-10
===================

  * feat(testdata): consider empty values as empty files

v0.1.0 / 2020-08-20
===================

  * feat: add subtests helper
  * docs(readme): typo
