bash apptron/scripts/run_phase2.sh > run_output.log 2>&1 &
PID=$!
echo "Waiting for wrangler to start..."
sleep 20
cd e2e_tests
npm ci || npm install
npx playwright install --with-deps chromium
npx playwright test
kill $PID
