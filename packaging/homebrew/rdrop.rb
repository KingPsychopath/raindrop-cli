class Rdrop < Formula
  desc "Find and organize Raindrop.io bookmarks with reviewed plans"
  homepage "https://github.com/KingPsychopath/raindrop-cli"
  url "https://github.com/KingPsychopath/raindrop-cli/archive/refs/tags/v0.2.1.tar.gz"
  sha256 "dff61f02b2750565fc7b0092bbb1e6c58e142b51f6793574fbefe3e881d09fee"
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
