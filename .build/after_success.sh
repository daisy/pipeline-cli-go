#!/bin/bash

if [ "$TRAVIS_BRANCH" == "master" ] && [ "$TRAVIS_PULL_REQUEST" == "false" ]; then
  make cover-deploy
  make dist
  mvn deploy -DskipTests -Dinvoker.skip=true --settings .maven/settings.xml
else
  echo "Skipping Maven deploy."
fi
