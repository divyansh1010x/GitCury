#!/bin/bash

# Files to update
files=(
    "/home/lakshya-jain/projects/GitCury/tests/core/core_test.go"
    "/home/lakshya-jain/projects/GitCury/tests/embeddings/embeddings_test.go"
    "/home/lakshya-jain/projects/GitCury/tests/git/git_test.go"
    "/home/lakshya-jain/projects/GitCury/tests/output/output_test.go"
    "/home/lakshya-jain/projects/GitCury/tests/utils/utils_test.go"
    "/home/lakshya-jain/projects/GitCury/tests/utils/custom_instructions_test.go"
)

for file in "${files[@]}"; do
    echo "Processing $file..."
    
    # Remove the duplicate SetupTestEnvironment function (from line with "SetupTestEnvironment" to line before first "func Test")
    # And replace calls to SetupTestEnvironment with testutils.SetupTestEnvironment
    sed -i '/\/\/ SetupTestEnvironment sets up environment variables for testing/,/^}$/d' "$file"
    sed -i 's/cleanup := SetupTestEnvironment(t)/cleanup := testutils.SetupTestEnvironment(t)/g' "$file"
    
    echo "Done processing $file"
done

echo "All files processed successfully!"
