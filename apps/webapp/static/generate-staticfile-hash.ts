import * as fs from "fs";
import * as crypto from "crypto";

function generateHash(filePath: string): string {
  const fileContent = fs.readFileSync(filePath, "utf8");
  const hash = crypto
    .createHash("sha256")
    .update(fileContent)
    .digest("hex")
    .substring(0, 16);
  return hash;
}

const arg = process.argv[2]; // Get the command-line argument

if (arg === "js") {
  const jsHash = generateHash("scripts.ts");
  console.log(`scripts-${jsHash}.js`);
} else if (arg === "css") {
  const cssHash = generateHash("app.css");
  console.log(`styles-${cssHash}.css`);
} else {
  console.error('Invalid argument. Use "js" or "css".');
  process.exit(1); // Exit with an error code
}
