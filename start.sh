#!/bin/bash

CLUSTER_FILE='Nodes.txt'
CLUSTER_START_FILE='main'
CLUSTER_KEY_FILE='clusterKey.txt'

MASTER_NODE_ID='Master'
PHILISOPH_NOD_ID='Philisoph'
MASTER_IP='127.0.0.1'
MASTER_PORT=8710
MASTER_HTTP_PORT=8001



function start_cluster_node () {
    if [[ "$1" == "init" ]] 
    then
        command gnome-terminal --tab --title=$2 -- /bin/bash -c \
            "./$CLUSTER_START_FILE $1 --node-name=$2 --bind-ip=$3 --bind-port=$4 --http-port=$5; bash"
        # command ./$CLUSTER_START_FILE init >output.txt 2>&1  &
    elif [[ "$1" == "join" ]] && [[ $# -eq 6 ]]
    then
        echo "joint startet: $5"
        command gnome-terminal --tab --title=$2  -- /bin/bash -c \
            "./$CLUSTER_START_FILE join \--node-name=$2 --known-ip=$3 \
            --cluster-key=$4 --bind-port=$5 --http-port=$6"
        # command ./$CLUSTER_START_FILE $1 --node-name=$2 --known-ip=$3 \
            #--cluster-key=$4 --bind-port=$5 --http-port=$6 >output.txt 2>&1  &
    fi
}   

function read_cluster_key() {

    if [[ $# -eq 1 ]]
    then
        n=1
        while IFS= read line; do
            echo $line
        done < $1       
    else
        return -1
    fi
}

function read_file_line() {
    if [[ $# -eq 1 ]]
    then
        n=1
        cluster_key=""
        cluster_port=""
        while read node_id node_add; do
            # reading each line
            #echo $node_id | egrep -o '^[^:]+'
            #take id until ":"
            nodeId=$(echo $node_id | cut -d: -f1)
            nodeIP=$(echo $node_add | sed 's/:.*//')
            #port-number after ":"
            nodePORT=$(echo $node_add | sed 's/.*:\(.*\)/\1/g')
            
            if [[ n -gt 1 ]]
            then
                
                start_cluster_node "join" "$nodeId" "127.0.0.1:$cluster_port" "$cluster_key" "$nodePORT" "800$n"
                echo "$nodeId startet"
                sleep 2
            else
                #First Member will take the Master Role
                start_cluster_node "init" "$nodeId" "$nodeIP" "$nodePORT" "800$n"
                echo "Master-Node startet in a new Tab!"

                readarray readFile <<< "$(read_cluster_key $CLUSTER_KEY_FILE)"
                cluster_key=${readFile[0]}
                cluster_port=${readFile[1]}
                #remove newline
                cluster_key="$(echo "$cluster_key"|tr -d '\n')"
                cluster_port="$(echo "$cluster_port"|tr -d '\n')"     
            fi
            n=$((n+1))
        done < $1
    else
        return -1
    fi
    
}

function start_MasterNode () {
    start_cluster_node "init" "$1" "$2" "$3" "$4"
    echo "Master-Node startet in a new Tab!"

}

function start_NNodes () {
    readarray readFile <<< "$(read_cluster_key $CLUSTER_KEY_FILE)"
    cluster_key=${readFile[0]}
    cluster_port=${readFile[1]}
    #remove newline
    cluster_key="$(echo "$cluster_key"|tr -d '\n')"
    cluster_port="$(echo "$cluster_port"|tr -d '\n')"


    philisoph_port=$MASTER_PORT
    philisoph_http_port=$MASTER_HTTP_PORT

    for (( i=1; i<=$1; i++ ))
    do  
        ((philisoph_port+=1))
        ((philisoph_http_port+=1))
        philisoph_node_id=$PHILISOPH_NOD_ID$i
        
        start_cluster_node "join" "$philisoph_node_id" "127.0.0.1:$cluster_port" "$cluster_key" "$philisoph_port" "$philisoph_http_port"
        # echo "$nodeId startet"
        # sleep 2
    done


}


function startNPhilosof () {
    if [[ $1 -gt 0 ]]
    then
        start_MasterNode $MASTER_NODE_ID $MASTER_IP $MASTER_PORT $MASTER_HTTP_PORT
        # start n nodes
        start_NNodes $1
    fi
}

function read_node_info_from_file () {
    if [[ ! -f "$CLUSTER_FILE" ]]
    then
        echo "file does not exist: $CLUSTER_FILE"
        exit 1
    else
        echo "inside"
    fi
}

function main () {
    echo "Cluster startet-------------------------------"
    read_file_line $CLUSTER_FILE
    # wait
}

# output=$(command ./main2 init &)
#coproc output { $(command ./main2 init >output.txt 2>&1  &) ;}
#coproc ./main2 init  &
#read -u ${output[0]} output1
#read -r output1 <&${output[0]}
#echo ${output_PID}
# command ./main2 init >output.txt 2>&1  &
# echo $output1


# start_cluster_node "init"
# main


command go build main.go
startNPhilosof $1

echo "All processes Runing!"
