[ -d log ] || mkdir log
logdir=log/$(date +%Y-%m-%d-%H-%M-%S)
[ -d $logdir ] || mkdir $logdir
echo "start db,log..."
./dbserver -c dbBase.json 1>$logdir/dbBase.log 2>$logdir/dbBase.err &
./dbserver -c dbExtern.json 1>$logdir/dbExtern.log 2>$logdir/dbExtern.err &
./logserver 1>$logdir/log.log 2>$logdir/log.err &
sleep 3
echo "start account,lock..."
./accountserver 1>$logdir/account.log 2>$logdir/account.err &
./lockserver 1>$logdir/lock.log 2>$logdir/lock.err &
sleep 3
echo "start center..."
./center 1>$logdir/center.log 2>$logdir/center.err &
sleep 30
echo "start chat,gate..."
./chatserver 1>$logdir/chat.log 2>$logdir/chat.err &
./gateserver 1>$logdir/gate.log 2>$logdir/gate.err &
sleep 3
echo "start gm,gas..."
./gmserver 1>$logdir/gm.log 2>$logdir/gm.err &
./gameserver -c gas1.json 1>$logdir/gas1.log 2>$logdir/gas1.err &
#./gameserver -c gas2.json 1>$logdir/gas2.log 2>$logdir/gas2.err &
#./gameserver -c gas3.json 1>$logdir/gas3.log 2>$logdir/gas3.err &
sleep 1
echo "start gmtools..."
../server/tools/GmTools/gmtools/gmtools 1>$logdir/gmtools.log 2>$logdir/gmtools.err &
sleep 1
echo "start all ok!"
