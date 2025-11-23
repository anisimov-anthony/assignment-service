#!/bin/bash
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'
SCENARIO=""
BASE_URL="${BASE_URL:-http://localhost:8080}"
DATE=$(date +%Y-%m-%d)

print_header() {
  echo -e "${BLUE}==========================================${NC}"
  echo -e "${BLUE}K6 Tests Runner${NC}"
  echo -e "${BLUE}Scenario: $SCENARIO${NC}"
  echo -e "${BLUE}Results Directory: k6-results/$DATE${NC}"
  echo -e "${BLUE}==========================================${NC}"
}

run_test() {
  local test_name=$1
  local script_path=$2
  local results_file="k6-results/$DATE/${test_name}_results.json"

  echo -e "${YELLOW}Running: $test_name${NC}"
  echo ""

  mkdir -p "k6-results/$DATE"

  if k6 run --out json="$results_file" "$script_path"; then
    echo -e "${GREEN}✅ Test completed: $test_name${NC}"
    return 0
  else
    echo -e "${RED}❌ Test failed: $test_name${NC}"
    return 1
  fi
}

run_sli_tests() {
  echo -e "${YELLOW}Running SLI tests...${NC}"
  echo ""

  local failed_tests=()
  local test_cases=(
    "sli_team_get:src/scenarios/sli/team_get/index.js"
    "sli_team_add:src/scenarios/sli/team_add/index.js"
    "sli_pr_create:src/scenarios/sli/pullRequest_create/index.js"
    "sli_pr_reassign:src/scenarios/sli/pullRequest_reassign/index.js"
    "sli_pr_merge:src/scenarios/sli/pullRequest_merge/index.js"
    "sli_users_getReview:src/scenarios/sli/users_getReview/index.js"
    "sli_users_setIsActive:src/scenarios/sli/users_setIsActive/index.js"
  )

  for test_case in "${test_cases[@]}"; do
    local test_name=$(echo "$test_case" | cut -d':' -f1)
    local script_path=$(echo "$test_case" | cut -d':' -f2)

    if ! run_test "$test_name" "$script_path"; then
      failed_tests+=("$test_name")
    fi

    sleep 2
    echo ""
  done

  if [ ${#failed_tests[@]} -eq 0 ]; then
    echo -e "${GREEN}All SLI tests completed successfully!${NC}"
  else
    echo -e "${RED}Failed tests: ${failed_tests[*]}${NC}"
    return 1
  fi
}

run_1k_tests() {
  echo -e "${YELLOW}Running 1k tests...${NC}"
  echo ""

  local failed_tests=()
  local test_cases=(
    "1k_team_get:src/scenarios/1k/team_get/index.js"
    "1k_team_add:src/scenarios/1k/team_add/index.js"
    "1k_pr_create:src/scenarios/1k/pullRequest_create/index.js"
    "1k_pr_reassign:src/scenarios/1k/pullRequest_reassign/index.js"
    "1k_pr_merge:src/scenarios/1k/pullRequest_merge/index.js"
    "1k_users_getReview:src/scenarios/1k/users_getReview/index.js"
    "1k_users_setIsActive:src/scenarios/1k/users_setIsActive/index.js"
  )

  for test_case in "${test_cases[@]}"; do
    local test_name=$(echo "$test_case" | cut -d':' -f1)
    local script_path=$(echo "$test_case" | cut -d':' -f2)

    if ! run_test "$test_name" "$script_path"; then
      failed_tests+=("$test_name")
    fi

    sleep 2
    echo ""
  done

  if [ ${#failed_tests[@]} -eq 0 ]; then
    echo -e "${GREEN}All 1k tests completed successfully!${NC}"
  else
    echo -e "${RED}Failed tests: ${failed_tests[*]}${NC}"
    return 1
  fi
}

run_all_tests() {
  echo -e "${YELLOW}Running all tests (SLI + 1k)...${NC}"
  echo ""

  run_sli_tests
  echo ""
  run_1k_tests
}

show_usage() {
  echo -e "${YELLOW}Usage: $0 [SCENARIO]${NC}"
  echo ""
  echo "Available scenarios:"
  echo "  sli                      Run all SLI tests (5 RPS)"
  echo "  1k                       Run all 1k tests (1000 RPS)"
  echo "  all                      Run all tests (SLI + 1k)"
  echo "  sli_team_get             Run specific SLI test"
  echo "  sli_team_add             Run specific SLI test"
  echo "  sli_pr_create            Run specific SLI test"
  echo "  sli_pr_reassign          Run specific SLI test"
  echo "  sli_pr_merge             Run specific SLI test"
  echo "  sli_users_getReview      Run specific SLI test"
  echo "  sli_users_setIsActive    Run specific SLI test"
  echo "  1k_team_get              Run specific 1k test"
  echo "  1k_team_add              Run specific 1k test"
  echo "  1k_pr_create             Run specific 1k test"
  echo "  1k_pr_reassign           Run specific 1k test"
  echo "  1k_pr_merge              Run specific 1k test"
  echo "  1k_users_getReview       Run specific 1k test"
  echo "  1k_users_setIsActive     Run specific 1k test"
  echo ""
  echo "Examples:"
  echo "  $0 sli"
  echo "  $0 1k"
  echo "  $0 all"
  echo "  $0 sli_pr_create"
}

mkdir -p k6-results

case "$1" in
sli)
  SCENARIO="sli"
  print_header
  run_sli_tests
  ;;
1k)
  SCENARIO="1k"
  print_header
  run_1k_tests
  ;;
all)
  SCENARIO="all"
  print_header
  run_all_tests
  ;;
sli_team_get)
  SCENARIO="sli_team_get"
  print_header
  run_test "sli_team_get" "src/scenarios/sli/team_get/index.js"
  ;;
sli_team_add)
  SCENARIO="sli_team_add"
  print_header
  run_test "sli_team_add" "src/scenarios/sli/team_add/index.js"
  ;;
sli_pr_create)
  SCENARIO="sli_pr_create"
  print_header
  run_test "sli_pr_create" "src/scenarios/sli/pullRequest_create/index.js"
  ;;
sli_pr_reassign)
  SCENARIO="sli_pr_reassign"
  print_header
  run_test "sli_pr_reassign" "src/scenarios/sli/pullRequest_reassign/index.js"
  ;;
sli_pr_merge)
  SCENARIO="sli_pr_merge"
  print_header
  run_test "sli_pr_merge" "src/scenarios/sli/pullRequest_merge/index.js"
  ;;
sli_users_getReview)
  SCENARIO="sli_users_getReview"
  print_header
  run_test "sli_users_getReview" "src/scenarios/sli/users_getReview/index.js"
  ;;
sli_users_setIsActive)
  SCENARIO="sli_users_setIsActive"
  print_header
  run_test "sli_users_setIsActive" "src/scenarios/sli/users_setIsActive/index.js"
  ;;
1k_team_get)
  SCENARIO="1k_team_get"
  print_header
  run_test "1k_team_get" "src/scenarios/1k/team_get/index.js"
  ;;
1k_team_add)
  SCENARIO="1k_team_add"
  print_header
  run_test "1k_team_add" "src/scenarios/1k/team_add/index.js"
  ;;
1k_pr_create)
  SCENARIO="1k_pr_create"
  print_header
  run_test "1k_pr_create" "src/scenarios/1k/pullRequest_create/index.js"
  ;;
1k_pr_reassign)
  SCENARIO="1k_pr_reassign"
  print_header
  run_test "1k_pr_reassign" "src/scenarios/1k/pullRequest_reassign/index.js"
  ;;
1k_pr_merge)
  SCENARIO="1k_pr_merge"
  print_header
  run_test "1k_pr_merge" "src/scenarios/1k/pullRequest_merge/index.js"
  ;;
1k_users_getReview)
  SCENARIO="1k_users_getReview"
  print_header
  run_test "1k_users_getReview" "src/scenarios/1k/users_getReview/index.js"
  ;;
1k_users_setIsActive)
  SCENARIO="1k_users_setIsActive"
  print_header
  run_test "1k_users_setIsActive" "src/scenarios/1k/users_setIsActive/index.js"
  ;;
*)
  show_usage
  exit 1
  ;;
esac
