class Rdrop < Formula
  desc "Find, review and organize Raindrop bookmarks from a shell or agent"
  homepage "https://github.com/KingPsychopath/raindrop-cli"
  url "https://github.com/KingPsychopath/raindrop-cli/archive/refs/tags/v0.2.0.tar.gz"
  sha256 "2be3fe887073288141ddc185fa7f639bea143b2275d5f1650c10b0f164130818"
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
