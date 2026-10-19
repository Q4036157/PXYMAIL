<?php
declare(strict_types=1);

$release = $argv[1] ?? 'E:/pxy-deploy/YXHK/instances/owner/current';
$destination = $argv[2] ?? 'E:/pxy-runtime/Mail/migration-data/yxhk-templates.json';
chdir($release);
$_SERVER['APP_ENV'] = $_ENV['APP_ENV'] = 'prod';
$_SERVER['APP_DEBUG'] = $_ENV['APP_DEBUG'] = '0';
putenv('APP_ENV=prod');
putenv('APP_DEBUG=0');
$app = require $release.'/app/console-application.php';
$app->getKernel()->boot();
$manager = $app->getKernel()->getContainer()->get('doctrine')->getManager();
$templates = [];
foreach ($manager->getRepository(\Mautic\EmailBundle\Entity\Email::class)->findAll() as $email) {
    $templates[] = [
        'id' => $email->getId(), 'name' => $email->getName(),
        'subject' => $email->getSubject(), 'html' => $email->getCustomHtml(),
        'text' => $email->getPlainText(), 'published' => $email->isPublished(false),
    ];
}
if (!is_dir(dirname($destination))) {
    mkdir(dirname($destination), 0700, true);
}
file_put_contents($destination, json_encode($templates, JSON_UNESCAPED_UNICODE | JSON_PRETTY_PRINT | JSON_THROW_ON_ERROR));
echo 'Exported '.count($templates)." templates.\n";
