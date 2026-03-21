import { execSync } from 'child_process';
import axios from 'axios';

export default async function globalSetup(config) {
  console.log('\n[Global Setup] Wiping Docker Postgres Test Entries...');
  
  try {
    // Purge the database to ensure 100% test idempotency
    execSync('docker exec unimatch_db psql -U postgres -d unimatch_be -c "DELETE FROM cases; DELETE FROM users;"', { stdio: 'inherit' });
    console.log('[Global Setup] Database wiped successfully.');
    
    // Auto-provision test accounts via Direct API call
    console.log('[Global Setup] Seeding test accounts...');
    await axios.post('http://localhost:8080/api/v1/auth/register', {
      username: 'e2e_admin',
      password: 'testpassword123'
    });
    console.log('[Global Setup] Root test account configured.');
  } catch (error) {
    console.error('[Global Setup] Failed during infrastructure wipe or seed:', error.message);
    throw error;
  }
}
