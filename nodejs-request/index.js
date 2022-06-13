'use strict';

const axios = require('axios');
const fsPromises = require('fs').promises;
const crypto = require('crypto');
const { spawn } = require('child_process');
const tar = require('tar');
const os = require('os');
const path = require('path');

const args = process.argv.slice(2);

const image = args[0];
const handler = args[1];
const lang = args[2];

async function build() {
  const tempDir = await fsPromises.mkdtemp(path.join(os.tmpdir(), 'builder-'));

  try {
    const tarFile = path.join(tempDir, 'req.tar');

    await shrinkwrap(image, handler, lang);

    let buildConfig = { image, buildArgs: {} };
    await fsPromises.writeFile(
      './build/com.openfaas.docker.config',
      JSON.stringify(buildConfig),
      'utf8'
    );

    await tar.c(
      {
        cwd: './build',
        file: tarFile,
      },
      ['./context', './com.openfaas.docker.config']
    );

    let secret = await fsPromises.readFile('./payload.txt', 'utf8');
    let data = await fsPromises.readFile(tarFile);
    let hash = crypto
      .createHmac('sha256', secret.trim())
      .update(data)
      .digest('hex');

    try {
      let res = await axios({
        data: data,
        method: 'post',
        url: 'http://127.0.0.1:8081/build',
        headers: {
          'Content-Type': 'application/octet-stream',
          'X-Build-Signature': 'sha256=' + hash,
        }
      })

      console.log(`Success building image ${res.data.image}`)
    } catch (err) {
      console.log(`Building image ${image} failed: ${err.response.data.status}`)
    }

  } catch (err) {
    throw err
  } finally {
    await fsPromises.rm(tempDir, { recursive: true });
  }
}

async function shrinkwrap(image, handler, lang) {
  const cmd = spawn('faas-cli', [
    'build',
    '--lang',
    lang,
    '--handler',
    handler,
    '--name',
    'context',
    '--image',
    image,
    '--shrinkwrap',
  ]);

  for await (const chunk of cmd.stdout) {
    console.log(chunk.toString());
  }

  let exitCode = await new Promise((resolve) => {
    cmd.on('close', resolve);
  });
  if (exitCode) {
    throw new Error('Failed to shrinkwrap handler');
  }
  return;
}

(async () => {
  try {
    await build();
  } catch (err) {
    console.error(err.message)
  }
})();
