package main

const bashCompletion = `# bash completion for rdrop
_rdrop_completion() {
  local cur prev commands
  COMPREPLY=()
  cur="${COMP_WORDS[COMP_CWORD]}"
  prev="${COMP_WORDS[COMP_CWORD-1]}"
  commands="mcp connect config context organize classify agent doctor me user stats collections collection covers cover export import upload tags tag filters cleanup list search untagged duplicates broken get add update delete batch highlights highlight cache sharing backups backup suggest completion raw version help"

  case "$prev" in
    rdrop)
      COMPREPLY=( $(compgen -W "$commands" -- "$cur") )
      return 0
      ;;
    completion)
      COMPREPLY=( $(compgen -W "bash zsh fish" -- "$cur") )
      return 0
      ;;
    mcp)
      COMPREPLY=( $(compgen -W "serve tools" -- "$cur") ); return 0 ;;
    connect)
      COMPREPLY=( $(compgen -W "mcp acp raindrop executor" -- "$cur") ); return 0 ;;
    config)
      COMPREPLY=( $(compgen -W "show paths" -- "$cur") ); return 0 ;;
    cleanup)
      COMPREPLY=( $(compgen -W "report plan apply" -- "$cur") ); return 0 ;;
    organize)
      COMPREPLY=( $(compgen -W "plan" -- "$cur") ); return 0 ;;
    collection)
      COMPREPLY=( $(compgen -W "create update delete sort expand merge clean empty-trash" -- "$cur") )
      return 0
      ;;
    tag)
      COMPREPLY=( $(compgen -W "rename merge delete" -- "$cur") )
      return 0
      ;;
    batch)
      COMPREPLY=( $(compgen -W "tag move delete" -- "$cur") )
      return 0
      ;;
  esac
}
complete -F _rdrop_completion rdrop
`

const zshCompletion = `#compdef rdrop
local -a commands
commands=(
  'mcp:serve MCP or inspect tools'
  'connect:print MCP ACP or remote configurations'
  'config:show redacted configuration'
  'context:export library context'
  'organize:validate model proposals'
  'classify:run a typed decision provider'
  'agent:run a direct model tool loop'
  'user:update user settings'
  'covers:search collection icons'
  'cover:upload cover'
  'upload:upload a file'
  'filters:read library counts'
  'highlights:list highlights'
  'highlight:manage highlights'
  'cache:download permanent copy'
  'sharing:manage collaborators'
  'backups:list backups'
  'backup:create or download backup'
  'suggest:get Raindrop suggestions'
  'doctor:check authentication'
  'me:show authenticated user'
  'stats:show account stats'
  'collections:list collections'
  'collection:manage collections'
  'tags:list tags'
  'tag:manage tags'
  'cleanup:build cleanup reports'
  'list:list bookmarks'
  'search:search bookmarks'
  'untagged:list untagged bookmarks'
  'duplicates:list duplicate bookmarks'
  'broken:list broken bookmarks'
  'get:get one bookmark'
  'add:add a bookmark'
  'update:update a bookmark'
  'delete:delete a bookmark'
  'batch:bulk bookmark operations'
  'export:export bookmarks'
  'import:import helpers'
  'raw:call a REST endpoint'
  'completion:print shell completion'
  'version:print version'
  'help:show help'
)
_describe 'command' commands
`

const fishCompletion = `# fish completion for rdrop
complete -c rdrop -f
complete -c rdrop -n "__fish_use_subcommand" -a "mcp connect config context organize classify agent user covers cover upload filters highlights highlight cache sharing backups backup suggest version help"
complete -c rdrop -n "__fish_seen_subcommand_from mcp" -a "serve tools"
complete -c rdrop -n "__fish_seen_subcommand_from connect" -a "mcp acp raindrop executor"
complete -c rdrop -n "__fish_seen_subcommand_from cleanup" -a "report plan apply"
complete -c rdrop -n "__fish_seen_subcommand_from organize" -a "plan"
complete -c rdrop -n "__fish_seen_subcommand_from config" -a "show paths"
complete -c rdrop -n "__fish_use_subcommand" -a "doctor" -d "Check authentication"
complete -c rdrop -n "__fish_use_subcommand" -a "me" -d "Show authenticated user"
complete -c rdrop -n "__fish_use_subcommand" -a "stats" -d "Show account stats"
complete -c rdrop -n "__fish_use_subcommand" -a "collections" -d "List collections"
complete -c rdrop -n "__fish_use_subcommand" -a "collection" -d "Manage collections"
complete -c rdrop -n "__fish_use_subcommand" -a "tags" -d "List tags"
complete -c rdrop -n "__fish_use_subcommand" -a "tag" -d "Manage tags"
complete -c rdrop -n "__fish_use_subcommand" -a "cleanup" -d "Build cleanup reports"
complete -c rdrop -n "__fish_use_subcommand" -a "list search untagged duplicates broken" -d "List bookmarks"
complete -c rdrop -n "__fish_use_subcommand" -a "get add update delete batch" -d "Modify bookmarks"
complete -c rdrop -n "__fish_use_subcommand" -a "export import raw" -d "Import, export, or call raw API"
complete -c rdrop -n "__fish_use_subcommand" -a "completion" -d "Print shell completion"
complete -c rdrop -n "__fish_seen_subcommand_from completion" -a "bash zsh fish"
`
