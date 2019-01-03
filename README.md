# JSON Response

<a href="https://circleci.com/gh/sylabs/workflows/json-resp"><img src="https://circleci.com/gh/sylabs/json-resp.svg?style=shield&circle-token=48bc85b347052f2de57405ab063b0d8b96c7059d"></a>
<a href="https://app.zenhub.com/workspace/o/sylabs/json-resp/boards"><img src="https://raw.githubusercontent.com/ZenHubIO/support/master/zenhub-badge.png"></a>

The `json-resp` package contains a small set of functions that are used to marshall and unmarshall response data and errors in JSON format.

## Quick Start

Install the [CircleCI Local CLI](https://circleci.com/docs/2.0/local-cli/). See the [Continuous Integration](#continuous-integration) section below for more detail.

To build and test:

```sh
circleci build
```

## Continuous Integration

This package uses [CircleCI](https://circleci.com) for Continuous Integration (CI). It runs automatically on commits and pull requests involving a protected branch. All CI checks must pass before a merge to a proected branch can be performed.

The CI checks are typically run in the cloud without user intervention. If desired, the CI checks can also be run locally using the [CircleCI Local CLI](https://circleci.com/docs/2.0/local-cli/).
