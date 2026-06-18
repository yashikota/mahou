#!/usr/bin/env bash
set -euo pipefail

target="${1:?target is required, e.g. darwin-arm64}"
out="${2:?output tar.zst path is required}"
root="${RUNNER_TEMP:-/tmp}/magickgo-runtime-${target}"

rm -rf "${root}"
mkdir -p "${root}"/{bin,lib,etc,share}

brew install imagemagick zstd ghostscript fontconfig || true
prefix="$(brew --prefix)"

cp -a "${prefix}/bin/magick" "${root}/bin/"
[ -x "${prefix}/bin/gs" ] && cp -a "${prefix}/bin/gs" "${root}/bin/" || true

cp -a "${prefix}"/lib/libMagickWand*.dylib "${root}/lib/"
cp -a "${prefix}"/lib/libMagickCore*.dylib "${root}/lib/"
find "${prefix}/lib" -maxdepth 1 -name '*.dylib' -exec cp -a {} "${root}/lib/" \;

for dir in "${prefix}"/lib/ImageMagick-*; do
  [ -d "${dir}" ] && cp -a "${dir}" "${root}/lib/"
done
for dir in "${prefix}"/etc/ImageMagick-*; do
  [ -d "${dir}" ] && cp -a "${dir}" "${root}/etc/"
done
for dir in "${prefix}"/share/ImageMagick-* "${prefix}"/share/fonts "${prefix}"/etc/fonts; do
  [ -d "${dir}" ] && cp -a "${dir}" "${root}/share/"
done

find "${root}/lib" -type f -name '*.dylib' | while read -r lib; do
  install_name_tool -id "@loader_path/$(basename "${lib}")" "${lib}" || true
  otool -L "${lib}" | awk '/\/opt\/homebrew|\/usr\/local/ {print $1}' | while read -r dep; do
    base="$(basename "${dep}")"
    [ -f "${root}/lib/${base}" ] && install_name_tool -change "${dep}" "@loader_path/${base}" "${lib}" || true
  done
done

if grep -R "/opt/homebrew\|/usr/local" "${root}/lib" >/dev/null 2>&1; then
  echo "absolute Homebrew references remain in runtime" >&2
  exit 1
fi

mkdir -p "$(dirname "${out}")"
tar -C "${root}" --zstd -cf "${out}" .

