#!/usr/bin/env bash
set -ex

targets=(
	testdata/repo/zip_directory/autod.zip
	testdata/repo/ref_zipped
	testdata/repo/zip_binary/autod.zip
)

for target in "${targets[@]}"; do
	sum=$(shasum -a 256 "${target}" | cut -d' ' -f1)
	echo "sum:${sum}"
	echo "target:${target}"
	grep -l -r "${target}?checksum=sha256" . | while IFS= read -r f; do
	  echo "updating:${f}"
	  gsed -i -e "s|${target}?checksum=sha256:[a-z0-9]{64}|${target}?checksum=sha256:${sum}|g" "${f}"
  done
done
