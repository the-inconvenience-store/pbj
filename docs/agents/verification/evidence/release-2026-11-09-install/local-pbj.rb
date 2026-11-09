class Pbj < Formula
  desc "Enriched macOS clipboard copy and paste commands"
  homepage "https://github.com/the-inconvenience-store/pbj"
  url "http://127.0.0.1:58153/pbj_0.0.0_source.tar.gz"
  version "0.0.0"
  sha256 "82c3ec1fe157c881450aac774749385dc649f7209af1a12932c9fb833ee72f95"

  depends_on "go" => :build
  depends_on macos: :ventura

  def install
    ENV["CGO_ENABLED"] = "1"
    ENV["GOTOOLCHAIN"] = "local"
    ENV["GOFLAGS"] = "-mod=readonly"
    ENV["MACOSX_DEPLOYMENT_TARGET"] = "13.0"
    ENV.append "CGO_CFLAGS", "-mmacosx-version-min=13.0"

    %w[pbjc pbjp].each do |command|
      system "go", "build", *std_go_args(output: bin/command, ldflags: "-X main.version=#{version}"), "./cmd/#{command}"
    end
  end

  test do
    %w[pbjc pbjp].each do |command|
      assert_equal "#{command} #{version}\n", shell_output("#{bin}/#{command} --version")
      assert_empty shell_output("#{bin}/#{command} --invalid-release-test-option", 2)
    end
  end
end
