import * as fs from "fs";
import * as crypto from "crypto";
import * as path from "path";

function generateHash(filePath: string): string {
  const fileContent = fs.readFileSync(filePath, "utf8");
  const hash = crypto
    .createHash("sha256")
    .update(fileContent)
    .digest("hex")
    .substring(0, 16);
  return hash;
}

const dirPath = process.argv[2];

if (!dirPath) {
  console.error("Usage: bun run generate-staticfile-hash <directory-path>");
  process.exit(1);
}

try {
  const files = fs.readdirSync(dirPath);

  files.forEach((file) => {
    const filePath = path.join(dirPath, file);
    const fileExt = path.extname(file);
    const fileBaseName = path.basename(file, fileExt);

    if (fileExt === ".js") {
      const jsHash = generateHash(filePath);
      const hashedJsFilename = `${fileBaseName}-${jsHash}.js`;
      const hashedJsPath = path.join(dirPath, hashedJsFilename);

      fs.renameSync(filePath, hashedJsPath);
      console.log(`Renamed: ${file} to ${hashedJsFilename}`);
    } else if (fileExt === ".css") {
      const cssHash = generateHash(filePath);
      const hashedCssFilename = `${fileBaseName}-${cssHash}.css`;
      const hashedCssPath = path.join(dirPath, hashedCssFilename);

      fs.renameSync(filePath, hashedCssPath);
      console.log(`Renamed: ${file} to ${hashedCssFilename}`);
    }
  });
} catch (err) {
  console.error("Error processing directory:", err);
  process.exit(1);
}
