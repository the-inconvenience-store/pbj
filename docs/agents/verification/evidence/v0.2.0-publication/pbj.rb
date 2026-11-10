class Pbj < Formula
  desc "Enriched macOS clipboard copy and paste commands"
  homepage "https://github.com/the-inconvenience-store/pbj"
  url "https://github.com/the-inconvenience-store/pbj/releases/download/v0.2.0/pbj_0.2.0_source.tar.gz"
  version "0.2.0"
  sha256 "3afb913a1ee9fff02ad0746cbf83282d49cca6d3bd9798967a9190565dafe3c7"

  depends_on "go" => :build
  depends_on macos: :ventura

  def install
    ENV["CGO_ENABLED"] = "1"
    ENV["GOTOOLCHAIN"] = "local"
    ENV["GOFLAGS"] = "-mod=readonly"
    ENV["MACOSX_DEPLOYMENT_TARGET"] = "13.0"
    ENV.append "CGO_CFLAGS", "-mmacosx-version-min=13.0"

    %w[pbjcopy pbjpaste].each do |command|
      system "go", "build", *std_go_args(output: bin/command, ldflags: "-X main.version=#{version}"), "./cmd/#{command}"
    end
  end

  test do
    %w[pbjcopy pbjpaste].each do |command|
      assert_equal "#{command} #{version}\n", shell_output("#{bin}/#{command} --version")
      assert_empty shell_output("#{bin}/#{command} --invalid-release-test-option", 2)
    end
  end
end
