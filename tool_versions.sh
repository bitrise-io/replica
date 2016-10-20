#!/bin/bash
set -ex

(
    vboxmanage --version
    vagrant version
    packer version
) &> tool_versions.log

cat tool_versions.log