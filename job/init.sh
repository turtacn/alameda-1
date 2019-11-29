#!/bin/sh
#

show_usage()
{
    /bin/echo -e "\n\nUsage: $0 [install|uninstall]\n\n"
    exit 1
}

#
# Main
#

# start crond only
case "$1" in
    "uninstall")
        /bin/echo -e "Start uninstalling...\n"
        /usr/local/bin/alameda-job uninstall        
        ;;
    "install")
        /bin/echo -e "Start installing...\n"
        /usr/local/bin/alameda-job install
        ;;
    *)
        show_usage
        exit $?
        ;;
esac

exit 0
