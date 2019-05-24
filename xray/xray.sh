#!/bin/bash
#
# vim: tabstop=4 shiftwidth=4 softtabstop=4
#
# Copyright (c) 2013-2019 ProphetStor Data Services, Inc.
# All Rights Reserved.
#
# - This program need to be executed in hosted-OS with enough available disk space, e.g. 10GB
# - This program need to execute "kubectl" command on k8s cluster
#
[ "${XRAYHOME}" = "" ] && XRAYHOME=/tmp

#
# Variables
APP_NAME="xray"
MAX_XRAYS=`echo ${MAX_XRAYS}`
[ "${MAX_XRAYS}" = "" ] && MAX_XRAYS=2  # keep maximum number of xray file

# TRAP exit
exit_code=1
on_exit()
{
    ret=$?
    release_exclusive_lock
    trap - EXIT # Disable exit handler
    exit ${ret}
}
# Assign exit handler
trap on_exit EXIT INT

show_usage()
{
    PRG="xray.sh"
    echo "Usage: ${PRG} [Option...]"
    echo
    echo "  Option:"
    echo "    -h,-?     show usage message"
    echo "    -v        verbose"
    echo "    -o        output file of X-ray"
    echo "    -k        keep working directory"
    echo
    echo "  Example:"
    echo "    ${PRG} -o xray-20140219162818.tgz"
    echo
}

# Find command tool in OS or alternative of it
prepare_os_commands()
{
    [ "`which timeout`" != "" ] && CMD_TIMEOUT="timeout --preserve-status -s KILL"
    [ "`which mkfs.ext4`" != "" ] && CMD_MKFS="mkfs.ext4 -j"
}

# get file
get_file()
{
    local time_out="$1"
    local src="$2"
    local dst="$3"

    [ "${CMD_TIMEOUT}" = "" ] && time_out=""
    retcode=0

    src=$(eval "echo \"${src}\"")  # expand environment variables to compute new src
    dst=${workdir}/$(eval "echo \"${dst}\"")  # expand environment variables to compute new dst

    i=`expr index "${src}" \*`
    m=$(echo "${src}" | wc -w)
    if [ "$i" -ne "0" -o "$m" -ge "2" ]
    then
        # ${src} is multiple, destination should be a directory
        # ex: copy '*.sh' to 'scripts'
        mkdir_dst=${dst}
    else
        # source is a single file
        echo "${dst}" | grep '/$' > /dev/null
        if [ "$?" -eq "0" ]
        then
            # ${dst} ends with '/', destination is a directory
            mkdir_dst=${dst}
        else
            # destination is a file
            mkdir_dst=$(dirname "${dst}")
        fi
    fi
    mkdir_ret=$(mkdir -p ${mkdir_dst} 2>&1 | grep "No space left on device")
    if [ "${mkdir_ret}" != "" ]
    then
        retcode=28
        exit 1
    fi

    if [ ${retcode} -eq 0 ]
    then
        if [ -d "${src}" -o -e "${src}" -o -e "`find ${src} 2> /dev/null | head -1`" ]; then
            echo "  copying ${src} to ${dst}"
            if [ "${verbose}" = "1" ]
            then
                ${CMD_TIMEOUT} ${time_out} sh -c "(cp -vH ${src} ${dst}) 2>&1"
            else
                ${CMD_TIMEOUT} ${time_out} sh -c "(cp -H ${src} ${dst} || true)" #> /dev/null 2>&1
            fi
        else
            echo " skip copying non-existing file(${src})"
        fi
        retcode=$?
    fi
    if [ ${retcode} -ne 0 ]
    then
        return ${retcode}
    fi

    echo

    return 0
}

# get output of command
get_command()
{
    local time_out=$1
    local outfile=$2
    local cmd="$3"

    # skip empty command
    [ "${cmd}" = "" ] && return 1
    [ "${CMD_TIMEOUT}" = "" ] && time_out=""

    # outfile filepath
    if [ "`echo \"${outfile}\" | cut -c 1`" != "/" ]
    then
        # relative path
        outfile="${workdir}/${outfile}"
    fi
    mkdir_dst=$(dirname "${outfile}")
    mkdir_ret=$(mkdir -p ${mkdir_dst} 2>&1 | grep "No space left on device")
    if [ "${mkdir_ret}" != "" ]
    then
        {retcode}=28
        exit ${retcode}
    fi
    echo "  executing \"${cmd}\" >> ${outfile}"
    echo "==================" >> ${outfile}
    echo "# ${cmd}" >> ${outfile}
    echo "==================" >> ${outfile}
    ${CMD_TIMEOUT} ${time_out} sh -c "(${cmd})" >> ${outfile} 2>&1
    retcode=$?
    echo "=== Return code is ${retcode} ==="
    echo >> ${outfile}
    echo >> ${outfile}

    echo
    return 0
}

acquire_exclusive_lock()
{
    workingroot=$1
    lockfile=${workingroot}/.xray.lck

    flock -x -w 300 357 >/dev/null 2>&1
    retcode=$?
    if [ ${retcode} -eq 65 ]
    then
        exec 357>${lockfile}
        flock -x -w 300 357 >/dev/null 2>&1
        retcode=$?
    fi
    if [ ${retcode} -ne 0 ]
    then
        return 16
    fi
}

release_exclusive_lock()
{
    flock -u 357 >/dev/null 2>&1
    exec 357>&-
}

prepare_working_dir()
{
    workingdir=$1
    workingroot=`dirname ${workingdir}`
    workingbox=${workingdir}.box

    # enough space?
    availexp=`stat -f -c "(%a*%s)/2" ${workingroot}`
    if [ "${availexp}" = "" ]
    then
        return 5
    fi
    availsize=$((${availexp}))
    if [ ${availsize} -le 52428800 ] # give up if available size is less than 50*2=100MB
    then
        return 28
    fi
    # exclusive lock
    acquire_exclusive_lock ${workingroot}
    retcode=$?
    if [ ${retcode} -ne 0 ]
    then
        return ${retcode}
    fi
    # prepare box
    mkdir ${workingdir} >/dev/null 2>&1
    dd if=/dev/zero of=${workingbox} bs=1 count=1 seek=${availsize} >/dev/null 2>&1
    echo y | ${CMD_MKFS} ${workingbox} >/dev/null 2>&1
    mount -o loop ${workingbox} ${workingdir} >/dev/null 2>&1
    retcode=$?
    if [ ${retcode} -ne 0 ]
    then
        umount -f ${workingdir} >/dev/null 2>&1
        rm -rf ${workingbox} ${workingdir} >/dev/null 2>&1
        return 5
    fi
}

clean_working_dir()
{
    workingroot=$1

    # exclusive lock
    acquire_exclusive_lock ${workingroot}
    retcode=$?
    if [ ${retcode} -ne 0 ]
    then
        return ${retcode}
    fi
    # clean box files
    boxfiles=`ls -1 ${workingroot}/*.box 2>/dev/null | xargs`
    for box in ${boxfiles}
    do
        umount -f ${box} >/dev/null 2>&1
        rm -rf ${box} ${box%*.*} >/dev/null 2>&1
    done
    # keep maximum MAX_XRAYS xrays
    MAX_XRAYS=`expr ${MAX_XRAYS} + 1`
    xrayfiles=`ls -1t ${workingroot}/xray_*.tgz* 2>/dev/null | tail -n +${MAX_XRAYS} | xargs`
    for xray in ${xrayfiles}
    do
        rm -f ${xray} >/dev/null 2>&1
    done
    release_exclusive_lock
}

print_error()
{
    error=$1
    echo
    if [ ${error} -eq 0 ]
    then
        echo "Success"
    elif [ ${error} -eq 28 ]
    then
        echo "No space left on device"
    elif [ ${error} -eq 16 ]
    then
        echo "Collecting x-ray is on going"
    elif [ ${error} -eq 5 ]
    then
        echo "I/O error"
    else
        echo "Error: ${error}"
    fi
    echo
}

#
# Main
#

# defaults
xraydir="xray_$(uname -n)_$(date +%Y%m%d%H%M%S)"
rootdir="${XRAYHOME}/var/tmp"
workdir="${rootdir}/${xraydir}"  # root directory of collecting files
export workdir  # variable can be inherited by child process
outfile="${rootdir}/${xraydir}.tgz"
keep=0
verbose=0

# find command
prepare_os_commands

# get arguments if supported
[ "`which getopts`" != "" ] && while getopts "h?vl:c:o:k" OPTION
do
    case "${OPTION}" in
        h|\?)
            show_usage
            exit 0
            ;;
        v)
            verbose=1
            ;;
        o)
            outfile=${OPTARG}
            ;;
        k)
            keep=1
            ;;
    esac
done

# clean up
[ ! -d ${rootdir} ] && mkdir -p ${rootdir}
clean_working_dir ${rootdir}

# prepare working space
prepare_working_dir ${workdir}
ret=$?
if [ ${ret} -ne 0 ]
then
    print_error ${ret}
    exit ${ret}
fi

if [ -e ${outfile} ]
then
    cnt=1
    while [ -e ${outfile%*.*}-${cnt}.tgz ]
    do
        cnt=`expr ${cnt} + 1`
    done
    outfile=${outfile%*.*}-${cnt}.tgz
fi
rm -f ${outfile}.temp

# prepare files
(
    # common
    get_file 180 '/etc/*release*' 'etc/'
    get_file 180 '/proc/cpuinfo' 'info/proc/'
    get_file 180 '/proc/meminfo' 'info/proc/'
    get_file 180 '/proc/buddyinfo' 'info/proc/'
    get_file 180 '/proc/slabinfo' 'info/proc/'
    get_file 180 '/etc/modprobe.d/*' 'info/etc_modprobe.d/'
    get_file 180 '/var/log/dmesg*' 'varlog/'
    get_file 180 '/var/log/kern.log' 'varlog/'
    # Redhat
    get_file 180 '/var/log/boot.log*' 'var/log'
    get_file 180 '/var/log/apport.log*' 'varlog/'
    get_file 180 '/var/log/messages' 'varlog/'
    get_file 180 '/var/log/ks-post.log' 'varlog/'
    get_file 180 '/var/log/installer/hardware-summary' 'varlog/installer/'
    get_file 180 '/var/log/installer/syslog' 'varlog/installer/'
    # Debian/Ubuntu
    get_file 180 '/var/log/syslog*' 'varlog/'
    get_file 180 '/etc/network/interfaces*' 'info/network/'
) | tee -a ${workdir}/xray.log

# prepare commands
(
    get_command 30 info/hwinfo "lspci -v -v -v"
    get_command 10 info/cpuinfo "cat /proc/cpuinfo"
    get_command 30 info/iptables "iptables -L -n -v"
    get_command 30 info/iptables "iptables -t nat -L -n -v"
    [ "`which ifconfig`" != "" ] && get_command 10 info/system "ifconfig -a"
    get_command 10 info/system "ip addr"
    get_command 10 info/system "hostname"
    get_command 10 info/system "uname -a"
    get_command 10 info/system "ps lawwwwx"
    get_command 30 info/system "df -l -x overlay | grep -v mounts/shm"
    get_command 120 info/systemctl "systemctl status -a -l --no-pager"
    [ "`which rpm`" != "" ] && get_command 30 info/pkgs-rpm "rpm -qa"
    [ "`which dpkg`" != "" ] && get_command 30 info/pkgs-deb "dpkg -l"
) | tee -a ${workdir}/xray.log

# k8s
(
    get_command 180 k8s/current-context "kubectl config current-context"
    get_command 180 k8s/version "kubectl version"
    get_command 180 k8s/nodes.yaml "kubectl get nodes -o yaml"
) | tee -a ${workdir}/xray.log

# alameda
(
    get_command 180 alameda/list "kubectl get pods --all-namespaces"
    get_command 180 alameda/log "kubectl get deployments --all-namespaces | egrep ' alameda-| admission-controller|federatorai-operator|prometheus' \
                                 | while read ns pod junk; do \
                                     echo \"kubectl -n \${ns} get deployment \${pod} -o yaml > ${workdir}/alameda/deployment.\${ns}.\${pod}.yaml\";\
                                   done | sh -x"
    get_command 300 alameda/log "kubectl get pods --all-namespaces | egrep ' alameda-| admission-controller|federatorai-operator|prometheus' \
                                 | while read ns pod junk; do \
                                     echo \"kubectl -n \${ns} get pod \${pod} -o yaml > ${workdir}/alameda/pod.\${ns}.\${pod}.yaml\";\
                                     echo \"kubectl -n \${ns} logs \${pod} > ${workdir}/alameda/\${ns}.\${pod}.log\";\
                                     echo \"kubectl -n \${ns} logs --previous \${pod} > ${workdir}/alameda/\${ns}.\${pod}.log.1\";\
                                   done | sh -x"
    get_command 300 alameda/log "kubectl get svc --all-namespaces | egrep ' alameda-|operator-admission-service|federatorai-operator|prometheus' \
                                 | while read ns pod junk; do \
                                     echo \"kubectl -n \${ns} get svc \${pod} -o yaml > ${workdir}/alameda/svc.\${ns}.\${pod}.yaml\";\
                                   done | sh -x"
     get_command 180 alameda/log "kubectl get configmaps --all-namespaces | grep federatorai-operator \
                                  | while read ns cm junk; do \
                                      echo \"kubectl -n \${ns} get configmap \${cm} -o yaml > ${workdir}/alameda/configmap.\${ns}.\${cm}.yaml\"; \
                                    done | sh -x"
    get_command 180 alameda/log "kubectl get crd | grep '^alameda' \
                                 | while read crd junk; do \
                                     echo \"kubectl get crd \${crd} -o yaml > ${workdir}/alameda/crd.\${crd}.yaml\"; \
                                   done | sh -x"
    get_command 180 alameda/log "kubectl get mutatingwebhookconfiguration | grep '^alameda' \
                                 | while read wh junk; do \
                                     echo \"kubectl get mutatingwebhookconfiguration -o yaml > ${workdir}/alameda/mutatingwebhookconfiguration.\${wh}.yaml\"; \
                                   done | sh -x"
    get_command 180 alameda/log "kubectl get validatingwebhookconfiguration | grep -v ^NAME \
                                 | while read wh junk; do \
                                     echo \"kubectl get validatingwebhookconfiguration -o yaml > ${workdir}/alameda/validatingwebhookconfiguration.\${wh}.yaml\"; \
                                   done | sh -x"
    get_command 180 alameda/alamedascalers.yaml "kubectl get alamedascalers --all-namespaces -o yaml"
    get_command 180 alameda/alamedarecommendations.yaml "kubectl get alamedarecommendations --all-namespaces -o yaml"
    get_command 180 alameda/alamedaservices.yaml "kubectl get alamedaservices --all-namespaces -o yaml"
    # copy /xray.sh from pod and run as "xray.sh [ns] [pod] [dest_dir]"
    # i.e each pod's xray.sh collect its files into <dest_dir>
    kubectl get pods --all-namespaces | egrep ' alameda-| admission-controller|federatorai-operator' \
        | while read ns pod junk; do \
            get_command 180 ${ns}.${pod}/log "dest_dir=${workdir}/${ns}.${pod}; \
            mkdir -p \${dest_dir}; \
            kubectl -n ${ns} cp ${pod}:/xray.sh \${dest_dir}/xray.sh; \
            sh -x \${dest_dir}/xray.sh ${ns} ${pod} \${dest_dir};"
          done
) | tee -a ${workdir}/xray.log

# create tar file
if [ "${verbose}" = "1" ]
then
    tar -zcvf ${outfile}.temp -C ${rootdir} ${xraydir}
else
    tar -zcvf ${outfile}.temp -C ${rootdir} ${xraydir} > /dev/null 2>&1
fi
mv -f ${outfile}.temp ${outfile}

echo "X-ray is generated as ${outfile}"
echo

# clean up
if [ "${keep}" -ne "1" ]
then
    clean_working_dir ${rootdir}
fi

exit ${ret}
