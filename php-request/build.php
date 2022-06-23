<?php

use Garden\Cli\Cli;
use Alchemy\Zippy\Zippy;

require 'vendor/autoload.php';

function shrinkwrap($image, $handler, $lang)
{
    $output=null;
    $retval=null;

    $faascli = array("faas-cli",
                     "build", 
                     "--lang", 
                     $lang,
                     "--handler",
                     $handler,
                     "--name",
                     "context",
                     "--image",
                     $image,
                     "--shrinkwrap"
                    );
 
    $cmd=implode(" ", $faascli);
    exec($cmd, $output, $retval);
    if($retval != 0) {
        throw new Exception("Failed to shrinkwrap handler");
      }
    
}

function os_path_join(...$parts) {
    return preg_replace('#'.DIRECTORY_SEPARATOR.'+#', DIRECTORY_SEPARATOR, implode(DIRECTORY_SEPARATOR, array_filter($parts)));
}

function makeTar($buildConfig, $path, $tarFile){
   
    $configFilePath = os_path_join($path, "com.openfaas.docker.config");
    $configFile = fopen($configFilePath, "w");
    fwrite($configFile, $buildConfig);
    fclose($configFile);

    $zippy = Zippy::load();
    $archive = $zippy->create($tarFile, array("./" => $path), true);
}

function callBuilder($tarFile){

$t = fopen($tarFile, "rb");
$data = fread($t, filesize($tarFile));
fclose($t);
$p = fopen('payload.txt', "rb");
$secret = trim(fread($p, filesize('payload.txt')));
fclose($p);
$digest = bin2hex(hash_hmac('sha256', $data, $secret, true));
$headers = ['X-Build-Signature' => "sha256=$digest",
            'Content-Type'      => 'application/octet-stream'
            ];
$client = new GuzzleHttp\Client();
$response = $client->post('http://127.0.0.1:8081/build', [
                          'body' => $data,
                          'headers' => $headers
                         ]);
return $response;
}

$cli = new Cli();

$cli->description('Build a function with the OpenFaaS Pro Builder')
    ->opt('image:i', 'Docker image name to build, e.g. docker.io/functions/hello-world:0.1.0', true)
    ->opt('handler:h', 'Directory with handler for function, e.g. ./hello-world', true)
    ->opt('lang:l', 'Language or template to use, e.g. node17', true);

// Parse and return cli args.
$args        = $cli->parse($argv, true);
$argsImage   = $args->getOpt('image');
$argsHandler = $args->getOpt('handler');
$argsLang    = $args->getOpt('lang');
 
$handlerPath = realpath($argsHandler);

$buildConfig = '{"image": "' .  $argsImage . '", "buildArgs" : {}}';

$tarFile = os_path_join(sys_get_temp_dir(), 'req.tar');

shrinkwrap($argsImage, $handlerPath, $argsLang);
makeTar($buildConfig, 'build', $tarFile);

$res = callBuilder($tarFile);

$statusCode = $res->getStatusCode();
$bodyArr = json_decode($res->getBody(),true);

if ($statusCode != 200) {
    print('Building image ' . $bodyArr['image'] . ' failed' . PHP_EOL);
    print($bodyArr['status'] . PHP_EOL);
}else{
    print('Success building image ' . $bodyArr['image'] . PHP_EOL);
}

if (file_exists($tarFile)) {
    unlink($tarFile);
 }

?>
