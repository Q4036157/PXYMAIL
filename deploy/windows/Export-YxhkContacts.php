<?php
declare(strict_types=1);
$release = 'E:/pxy-deploy/YXHK/instances/owner/current';
chdir($release);
$_SERVER['APP_ENV'] = $_ENV['APP_ENV'] = 'prod';
$_SERVER['APP_DEBUG'] = $_ENV['APP_DEBUG'] = '0';
putenv('APP_ENV=prod');
putenv('APP_DEBUG=0');
$app = require $release.'/app/console-application.php';
$app->getKernel()->boot();
$db = $app->getKernel()->getContainer()->get('doctrine')->getConnection();
$contacts = $db->executeQuery('SELECT id,email,firstname,lastname FROM leads WHERE email IS NOT NULL AND email <> \'\' ORDER BY id')->fetchAllAssociative();
$dnc = $db->executeQuery("SELECT lead_id,reason,comments FROM lead_donotcontact WHERE channel='email'")->fetchAllAssociative();
foreach ($contacts as &$contact) { $contact['id'] = (int)$contact['id']; }
unset($contact);
foreach ($dnc as &$entry) { $entry['lead_id'] = (int)$entry['lead_id']; $entry['reason'] = (int)$entry['reason']; }
unset($entry);
file_put_contents('E:/pxy-runtime/Mail/migration-data/yxhk-contacts.json', json_encode(['contacts'=>$contacts,'do_not_contact'=>$dnc], JSON_UNESCAPED_UNICODE|JSON_THROW_ON_ERROR));
echo 'Exported '.count($contacts).' contacts and '.count($dnc)." exclusion records.\n";
