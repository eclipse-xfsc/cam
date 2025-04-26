# Contributing Guidelines

Please adhere to the following agreed contributing guidelines:

## Branches and Merge Requests

* The development of the CAM is conducted using branches and merge requests.
* Open a new branch for each new feature or change and try to keep the changes limited in size.
* Give your branch a short descriptive name and use a pre-fix if the change is focused on one dedicated part/module of the CAM, e.g., `collection/integrity` for work on the collection module for remote integrity.
* Use clear and speaking commit messages. You may use a prefix to illustrate which files/parts you were working on, e.g. `README: added default ports`
* When your work in branch is finalized and the CI/CD pipeline works for the changes, you can open a merge request to the get the branch merged into the main branch. Provide a short description and link any issues that might be closed by this pull request.

## Go Packages
To reduce dependencies, the different modules/parts of the CAM shall use the same set of packages for common tasks:

* Logging: logrus (https://github.com/sirupsen/logrus)
* Regular triggering of actions: Gocron (https://github.com/go-co-op/gocron)

## Licensing
If existing code is reused for this repo, the license for this code must be compatible with Apache 2.0.
