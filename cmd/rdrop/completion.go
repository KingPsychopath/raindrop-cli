package main

const bashCompletion = `# bash completion for rdrop
_rdrop_completion() {
  local cur prev commands
  COMPREPLY=()
  cur="${COMP_WORDS[COMP_CWORD]}"
  prev="${COMP_WORDS[COMP_CWORD-1]}"
  commands="doctor me user stats collections collection covers cover export import upload tags tag filters cleanup list search untagged duplicates broken get add update delete batch highlights highlight cache sharing backups backup suggest completion raw version help"

  case "$prev" in
    rdrop)
      COMPREPLY=( $(compgen -W "$commands" -- "$cur") )
      return 0
      ;;
    completion)
      COMPREPLY=( $(compgen -W "bash zsh fish" -- "$cur") )
      return 0
      ;;
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
