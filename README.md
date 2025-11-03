# `quietus`

[![tests](https://github.com/ianlewis/quietus/actions/workflows/pull_request.tests.yml/badge.svg)](https://github.com/ianlewis/quietus/actions/workflows/pull_request.tests.yml)
[![Codecov](https://codecov.io/gh/ianlewis/quietus/graph/badge.svg?token=Y0BLEHUAEC)](https://codecov.io/gh/ianlewis/quietus)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/ianlewis/quietus/badge)](https://securityscorecards.dev/viewer/?uri=github.com%2Fianlewis%2Fquietus)

`quietus` is a CLI tool to send signals to applications. It can be used to
gracefully terminate processes by sending a `SIGTERM` then `SIGKILL` after a
timeout if a process hasn't terminated on its own.
