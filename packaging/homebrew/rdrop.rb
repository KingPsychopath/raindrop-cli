class Rdrop < Formula
  desc "Dependency-light CLI for managing Raindrop.io libraries"
  homepage "https://github.com/KingPsychopath/raindrop-cli"
  url "https://github.com/KingPsychopath/raindrop-cli/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "b53eff457c3c72cdfa553bd052fdad2ca74806798e75e5bd1369defac01b3457"
  license "MIT"
  head "https://github.com/KingPsychopath/raindrop-cli.git", branch: "main"

  depends_on "go" => :build

  def install
    ldflags = %W[
      -s -w
      -X main.version=#{version}
      -X main.commit=homebrew
      -X main.date=unknown
    ]
    system "go", "build", "-trimpath", "-ldflags=#{ldflags.join(" ")}", "-o", bin/"rdrop", "./cmd/rdrop"
    generate_completions_from_executable(bin/"rdrop", "completion")
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/rdrop version")
    assert_match "Usage:", shell_output("#{bin}/rdrop help")
  end
end
