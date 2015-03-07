<?php
$socket = socket_create(AF_INET, SOCK_STREAM, SOL_TCP);
socket_set_option($socket, SOL_SOCKET, SO_REUSEADDR, 1);
socket_set_option($socket, SOL_SOCKET, SO_KEEPALIVE, 10000);
socket_set_option($socket, SOL_SOCKET, SO_RCVTIMEO, array("sec"=>5, "usec"=>0));
$con=socket_connect($socket,'127.0.0.1',8280);
if(!$con){socket_close($socket);exit;}
echo "Link\n";

$auth = array(
		'Cmd' => "AUTH",
		'Params' => "gim#test#key&gim#test#key#secret&Jack&123456",
	);

$str = json_encode($auth) . "\n";
$len = socket_write($socket, $str, strlen($str));
echo $len . "\n";

while($hear=socket_read($socket,1024)){
		
		// $len = socket_write($socket, $str, strlen($str));
		// echo $len . "\n";
        echo $hear;
}

socket_close($socket);
