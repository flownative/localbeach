# frozen_string_literal: true

#
# DO NOT EDIT THIS FILE MANUALLY
#
class Localbeach < Formula
  desc "Command-line tool for Flownative Beach"
  homepage "https://www.flownative.com/localbeach"
  version "{{VERSION}}"
  license "GPL-3.0-or-later"

  on_macos do
    on_intel do
      url "https://github.com/flownative/localbeach/releases/download/{{VERSION}}/beach_darwin_amd64.zip"
      sha256 "{{DARWIN_AMD64_SHA256SUM}}"
    end
    on_arm do
      url "https://github.com/flownative/localbeach/releases/download/{{VERSION}}/beach_darwin_arm64.zip"
      sha256 "{{DARWIN_ARM64_SHA256SUM}}"
    end
  end

  on_linux do
    url "https://github.com/flownative/localbeach/releases/download/{{VERSION}}/beach_linux_amd64.zip"
    sha256 "{{LINUX_SHA256SUM}}"
  end

  depends_on "mkcert"
  depends_on "nss"

  def install
    bin.install "beach" => "beach"
  end

  def caveats
    <<~EOS
    Local Beach is built on top of Docker and Docker Compose.
    You will need a working setup of both in order to use Local Beach.
    EOS
  end
end
