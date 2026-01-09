
<a name="v0.2.7"></a>
## [v0.2.7](https://github.com/xoctopus/confx/compare/v0.2.6...v0.2.7)

> 2026-01-07

### Doc

* update components documents
* update changelog


<a name="v0.2.6"></a>
## [v0.2.6](https://github.com/xoctopus/confx/compare/v0.2.5...v0.2.6)

> 2026-01-07

### Ci

* update Makefile

### Doc

* update changelog

### Feat

* async call with once callback
* **confpulsar:** producer options
* **types:** universal Endpoint with custom Option and LivenessChecker

### Fix

* close client if init failed

### Test

* **confpulsar:** fix pulsar liveness check and optimize testing


<a name="v0.2.5"></a>
## [v0.2.5](https://github.com/xoctopus/confx/compare/v0.2.4...v0.2.5)

> 2026-01-05

### Build

* **deps:** bump codecov/codecov-action from 4 to 5 ([#1](https://github.com/xoctopus/confx/issues/1))
* **deps:** bump actions/setup-go from 5 to 6 ([#2](https://github.com/xoctopus/confx/issues/2))
* **deps:** bump actions/checkout from 4 to 6 ([#3](https://github.com/xoctopus/confx/issues/3))
* **deps:** bump github.com/xoctopus/x from 0.2.10 to 0.2.11 ([#4](https://github.com/xoctopus/confx/issues/4))
* **deps:** bump github.com/xoctopus/genx from 0.1.14 to 0.1.15 ([#5](https://github.com/xoctopus/confx/issues/5))

### Feat

* **confpulsar:** endpoint options
* **types:** duration for url values


<a name="v0.2.4"></a>
## [v0.2.4](https://github.com/xoctopus/confx/compare/v0.2.3...v0.2.4)

> 2026-01-04

### Chore

* update dependencies
* update dependencies


<a name="v0.2.3"></a>
## [v0.2.3](https://github.com/xoctopus/confx/compare/v0.2.2...v0.2.3)

> 2026-01-04

### Chore

* rename injector

### Ci

* fix hack and Makefile

### Feat

* **appx:** appx.Conf with context
* **cmdx:** flag no option default value

### Fix

* endpoint retrier
* **cmdx:** NoOptionDefVal

### Test

* fix unit test


<a name="v0.2.2"></a>
## [v0.2.2](https://github.com/xoctopus/confx/compare/v0.2.1...v0.2.2)

> 2026-01-01

### Ci

* update ci.yml


<a name="v0.2.1"></a>
## [v0.2.1](https://github.com/xoctopus/confx/compare/v0.2.0...v0.2.1)

> 2026-01-01

### Feat

* **cmdx:** easy cli flag parser


<a name="v0.2.0"></a>
## [v0.2.0](https://github.com/xoctopus/confx/compare/v0.1.7...v0.2.0)

> 2025-12-25

### Chore

* **envconf:** code formatted

### Feat

* **envconf:** stronger validation

### Refact

* envconf and confapp

### Test

* **confmws:** fix unit test
* **envconf:** remove testdata import


<a name="v0.1.7"></a>
## [v0.1.7](https://github.com/xoctopus/confx/compare/v0.1.6...v0.1.7)

> 2025-02-10

### Chore

* add test depends files
* cleanup temporory files
* **confcmd:** update example


<a name="v0.1.6"></a>
## [v0.1.6](https://github.com/xoctopus/confx/compare/v0.1.5...v0.1.6)

> 2024-07-31


<a name="v0.1.5"></a>
## [v0.1.5](https://github.com/xoctopus/confx/compare/v0.1.4...v0.1.5)

> 2024-07-19


<a name="v0.1.4"></a>
## [v0.1.4](https://github.com/xoctopus/confx/compare/v0.1.3...v0.1.4)

> 2024-07-19


<a name="v0.1.3"></a>
## [v0.1.3](https://github.com/xoctopus/confx/compare/v0.1.2...v0.1.3)

> 2024-07-19


<a name="v0.1.2"></a>
## [v0.1.2](https://github.com/xoctopus/confx/compare/v0.1.1...v0.1.2)

> 2024-07-19

### Fix

* **confapp:** fix app meta overwrite `DefaultMeta`

### Refactor

* **confcmd:** refactor confcmd


<a name="v0.1.1"></a>
## [v0.1.1](https://github.com/xoctopus/confx/compare/v0.1.0...v0.1.1)

> 2024-07-09

### Feat

* **confapp:** gen default when app configured, fix masked string


<a name="v0.1.0"></a>
## [v0.1.0](https://github.com/xoctopus/confx/compare/v0.0.8...v0.1.0)

> 2024-07-09

### Fix

* **envconf:** fix slice and map en(de)coder and add constraints of map key


<a name="v0.0.8"></a>
## [v0.0.8](https://github.com/xoctopus/confx/compare/v0.0.7...v0.0.8)

> 2024-06-13

### Docs

* update README

### Test

* update flagValue unit test


<a name="v0.0.7"></a>
## [v0.0.7](https://github.com/xoctopus/confx/compare/v0.0.6...v0.0.7)

> 2024-06-02

### Docs

* update README

### Feat

* **confcmd:** add flag options: persistent, shorthand and no option default value
* **confcmd:** update Executor interface

### Test

* add unit test

### BREAKING CHANGE


`Executor` interface changed


<a name="v0.0.6"></a>
## [v0.0.6](https://github.com/xoctopus/confx/compare/v0.0.5...v0.0.6)

> 2024-06-01

### Feat

* **confcmd:** flags can be parsed by `encoding.TextUnmarshaller/TextMarshaller` if it was impled


<a name="v0.0.5"></a>
## [v0.0.5](https://github.com/xoctopus/confx/compare/v0.0.4...v0.0.5)

> 2024-05-29

### Feat

* **confcmd:** use env var to cover default value


<a name="v0.0.4"></a>
## [v0.0.4](https://github.com/xoctopus/confx/compare/v0.0.3...v0.0.4)

> 2024-05-29

### Build

* makefile add code format entry

### Refactor

* **confcmd:** rewrite executor interface and add unit test


<a name="v0.0.3"></a>
## v0.0.3

> 2024-05-28

### Build

* xgo experiment

### Chore

* clean build file add gitignore
* **confapp:** output more panic information

### Ci

* modify github workflow for triggering ci on tags
* add github ci workflow

### Docs

* add README
* add license

### Feat

* upgrade x v0.0.9=>v0.0.10
* update go mod
* upgrade go modules
* **commands:** defines commands for generate ci, config and workflow files
* **confapp:** complete confapp dev and add an example
* **confapp:** add example for confapp
* **confapp:** impls universal go command app (wip)
* **confcmd:** command facotry from struct value
* **envconf:** add `Var` option for more var description
* **envconf:** add var group marshaling methods
* **envconf:** add envconf for encoding or decoding dotenv

### Refactor

* rename module, migrated to xoctopus

### Test

* fix confapp unit test
* **confapp:** fix unit test

