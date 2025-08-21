set -e

go build -o ast app/*.go

echo "--- Running Local Tests ---"

# --- Run test 1: Match literal character ---
echo "Match literal character"
echo -n "apple" | ./ast -E "a"
echo "Test 1 passed."
echo ""

# --- Run test 2: Match digits ---
echo "Match digits"
echo -n "apple123" | ./ast -E "\d"
echo "Test 2 passed."
echo ""

# --- Run test 3: Match alphanumeric characters ---
echo "Match alphanumeric characters"
echo -n "alpha-num3ric" | ./ast -E "\w"
echo "Test 3 passed."
echo ""

# --- Run test 4: Positive character groups ---
echo "Positive character groups"
echo -n "apple" | ./ast -E "[abc]"
echo "Test 4 passed."
echo ""

# --- Run test 5: Negative character groups ---
echo "Negative character groups"
echo -n "apple" | ./ast -E "[^abc]"
echo -n "apple" | ./ast -E "[^bcd]"
echo "Test 5 passed."
echo ""

# --- Run test 6: Combining character groups ---
echo "Combining character groups"
echo -n "1 apple" | ./ast -E "\d apple"
echo "Test 6 passed."
echo ""

rm ast