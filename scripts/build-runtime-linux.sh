#!/usr/bin/env bash
set -euo pipefail

target="${1:?target is required, e.g. linux-amd64}"
out="${2:?output tar.zst path is required}"
root="${RUNNER_TEMP:-/tmp}/magickgo-runtime-${target}"
build="${RUNNER_TEMP:-/tmp}/magickgo-build-${target}"
prefix="${build}/prefix"
imagemagick_version="${IMAGEMAGICK_VERSION:-7.1.2-8}"
imagemagick_url="${IMAGEMAGICK_URL:-https://imagemagick.org/archive/releases/ImageMagick-${imagemagick_version}.tar.xz}"

rm -rf "${root}" "${build}"
mkdir -p "${root}"/{bin,lib,etc,share}
mkdir -p "${build}" "${prefix}"

sudo apt-get update
sudo apt-get install -y \
  autoconf \
  automake \
  build-essential \
  curl \
  fontconfig \
  ghostscript \
  libbz2-dev \
  libdjvulibre-dev \
  libfftw3-dev \
  libfontconfig1-dev \
  libfreetype6-dev \
  libheif-dev \
  libjbig-dev \
  libjpeg-dev \
  libjxl-dev \
  liblcms2-dev \
  liblqr-1-0-dev \
  libltdl-dev \
  liblzma-dev \
  libopenexr-dev \
  libopenjp2-7-dev \
  libpango1.0-dev \
  libpng-dev \
  libraqm-dev \
  librsvg2-dev \
  libraw-dev \
  libtiff-dev \
  libtool \
  libwebp-dev \
  libwmf-dev \
  libxml2-dev \
  libzip-dev \
  patchelf \
  pkg-config \
  tar \
  xz-utils \
  zlib1g-dev \
  zstd

curl -fsSL "${imagemagick_url}" -o "${build}/imagemagick.tar.xz"
tar -C "${build}" --strip-components=1 -xJf "${build}/imagemagick.tar.xz"

(
  cd "${build}"
  ./configure \
    --prefix="${prefix}" \
    --disable-static \
    --enable-shared \
    --with-modules \
    --with-quantum-depth=16 \
    --enable-hdri \
    --with-bzlib \
    --with-djvu \
    --with-fftw \
    --with-fontconfig \
    --with-freetype \
    --with-heic \
    --with-jbig \
    --with-jpeg \
    --with-jxl \
    --with-lcms \
    --with-lqr \
    --with-lzma \
    --with-openexr \
    --with-openjp2 \
    --with-pango \
    --with-png \
    --with-raqm \
    --with-raw \
    --with-rsvg \
    --with-tiff \
    --with-webp \
    --with-wmf \
    --with-xml \
    --with-zlib \
    --with-zstd
  make -j"$(nproc)"
  make install
)

cp -a "${prefix}/bin/magick" "${root}/bin/"
if command -v gs >/dev/null 2>&1; then
  cp -a "$(command -v gs)" "${root}/bin/"
fi

cp -a "${prefix}"/lib/libMagickWand*.so* "${root}/lib/"
cp -a "${prefix}"/lib/libMagickCore*.so* "${root}/lib/"
for dir in "${prefix}"/lib/ImageMagick-*; do
  [ -d "${dir}" ] && cp -a "${dir}" "${root}/lib/"
done
for dir in "${prefix}"/etc/ImageMagick-*; do
  [ -d "${dir}" ] && cp -a "${dir}" "${root}/etc/"
done
for dir in "${prefix}"/share/ImageMagick-* /usr/share/fonts /usr/share/fontconfig; do
  [ -d "${dir}" ] && cp -a "${dir}" "${root}/share/"
done

copy_deps() {
  local changed=1
  while [ "${changed}" -eq 1 ]; do
    changed=0
    while read -r lib; do
      [ -f "${lib}" ] || continue
      name="$(basename "${lib}")"
      case "${name}" in
        libc.so.*|libpthread.so.*|libm.so.*|libdl.so.*|librt.so.*|libgcc_s.so.*|libstdc++.so.*|libresolv.so.*|ld-linux-*.so.*|libutil.so.*|libcrypt.so.*)
          continue
          ;;
      esac
      dest="${root}/lib/${name}"
      if [ ! -e "${dest}" ]; then
        cp -a "${lib}" "${dest}" || true
        changed=1
      fi
    done < <(
      find "${root}/bin" "${root}/lib" -type f \
        \( -perm -0100 -o -name '*.so*' \) -print0 |
        xargs -0 -r env LD_LIBRARY_PATH="${root}/lib:${prefix}/lib" ldd 2>/dev/null |
        awk '/=> \// {print $3} /^\// {print $1}' |
        sort -u
    )
  done
}
copy_deps

find "${root}/bin" "${root}/lib" -type f \( -perm -0100 -o -name '*.so*' \) \
  -exec patchelf --set-rpath '$ORIGIN/../lib:$ORIGIN' {} \; 2>/dev/null || true

policy_dir="$(find "${root}/etc" -maxdepth 1 -type d -name 'ImageMagick-*' | sort | tail -n 1)"
[ -n "${policy_dir}" ] || policy_dir="${root}/etc/ImageMagick-7"
mkdir -p "${policy_dir}"
cat >"${policy_dir}/policy.xml" <<'XML'
<?xml version="1.0" encoding="UTF-8"?>
<policymap>
  <policy domain="coder" rights="none" pattern="PDF" />
  <policy domain="coder" rights="none" pattern="PS" />
  <policy domain="coder" rights="none" pattern="EPS" />
  <policy domain="coder" rights="none" pattern="MVG" />
  <policy domain="coder" rights="none" pattern="MSL" />
  <policy domain="delegate" rights="none" pattern="URL" />
  <policy domain="delegate" rights="none" pattern="HTTP" />
  <policy domain="delegate" rights="none" pattern="HTTPS" />
</policymap>
XML

env LD_LIBRARY_PATH="${root}/lib" "${root}/bin/magick" -version
env LD_LIBRARY_PATH="${root}/lib" "${root}/bin/magick" -list configure | grep -E 'QuantumDepth|HDRI' || true

mkdir -p "$(dirname "${out}")"
tar -C "${root}" --zstd -cf "${out}" .
