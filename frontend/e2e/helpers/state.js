import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const stateDir = path.resolve(__dirname, '../../test-results');
const stateFile = path.join(stateDir, 'e2e-state.json');

export async function writeRunState(payload) {
  await fs.mkdir(stateDir, { recursive: true });
  await fs.writeFile(stateFile, JSON.stringify(payload, null, 2), 'utf8');
}

export async function readRunState() {
  const raw = await fs.readFile(stateFile, 'utf8');
  return JSON.parse(raw);
}

export function getRunStatePath() {
  return stateFile;
}
