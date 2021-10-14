#!/bin/bash -e

Help() {
  # Display Help
  echo "Notify slack channel with a message."
  echo
  echo "Syntax: $0 [-c|m|w|h]"
  echo "options:"
  printf "\t-c     Slack channel to notify.\n"
  printf "\t-m     Use this message for slack notification (use - to read from stdin).\n"
  printf "\t-w     Slack webhook URL to send message to.\n"
  printf "\t-h     Print this help.\n"
  echo
}

main() {
  cd "$(dirname "$0")"
  local CHANNEL_ID=
  local MESSAGE=
  local SLACK_WEBHOOK_URL=
  while getopts ":hc:m:w:" option; do
    case $option in
    h) # display Help
      Help
      exit
      ;;
    c) # Channel to notify
      CHANNEL_ID=$OPTARG
      ;;
    m) # Use this message for slack notification
      MESSAGE=$OPTARG
      ;;
    w) # Use this message for slack notification
      SLACK_WEBHOOK_URL=$OPTARG
      ;;
    \?) # Invalid option
      echo "Error: Invalid option -$OPTARG. Use -h to see valid options."
      exit 1
      ;;
    esac
  done

  if [ ! "$CHANNEL_ID" ] || [ ! "$MESSAGE" ] || [ ! "$SLACK_WEBHOOK_URL" ]; then
    echo "Need to specify all options: channel ID (-c), message (-m) and slack webhook URL (-w). Use -h to see valid options."
    exit 1
  fi

  if [ "$MESSAGE" = "-" ]; then
    MESSAGE=$(cat /dev/stdin)
  fi


  curl -X POST --data-urlencode "payload={\"channel\": \"${CHANNEL_ID}\", \"username\": \"jenkins\", \"text\": \"${MESSAGE}\"}" "${SLACK_WEBHOOK_URL}"

}

main "$@"
