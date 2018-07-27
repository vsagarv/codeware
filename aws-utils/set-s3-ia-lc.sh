#!/bin/bash

progname=`basename $0`

# -----------------------------------------------------------------------------------
usage() {
	cat <<USAGE_END
usage: $progname -p <AWS CLI profile> -r <AWS region> [-d <days>]

# reads a list of S3 bucket names - one per line - from stdin
# -p: AWS CLI profile (--profile arg of aws command)
# -r: AWS region of the S3 buckets
# -d: Number of days (>= 30) after which an asset has to move from S3 to S3-IA (default = 30)
USAGE_END
}

if [ $# -lt 3 ]
then
	usage
	exit 1
fi

# -----------------------------------------------------------------------------------
check_and_set_lifecycle_config() {
	# for each bucket name provided on stdin, check if it has a lifecycle config set and if not, set an S3->IA lifecycle config
	while read bucket_name
	do
		[ -n "$bucket_name" ] || continue

		echo "looking for any existing lifecycle configs for $bucket_name bucket ..."	

		lc_cfg=`aws --profile=${cli_profile} s3api get-bucket-lifecycle --bucket ${bucket_name} 2>&1`

		echo "Current lifecycle configuration:"
		echo "$lc_cfg"

		if [[ $lc_cfg = *NoSuchBucket* || $lc_cfg = *AccessDenied* || $lc_cfg = *AllAccessDisabled* ]]
		then
			false # nop
		elif [[ $lc_cfg = *NoSuchLifecycleConfiguration* && $lc_cfg != *Transition* && $lc_cfg != *Expiration* ]]
		then
			echo "none found; setting one for $days days S3->IA transition like this from $lc_json_work_fname:"
			echo "###"
			cat $lc_json_work_fname
			echo "###"

			echo "sure? [yes/no]"
			while read yes_no </dev/tty
			do
				echo "yes_no: $yes_no"
				[ "$yes_no" = "yes" -o "$yes_no" = "no" ] && break
			done

			if [ "$yes_no" = "no" ]
			then
				echo "phew! the wise are cautious."
			elif [ "$yes_no" = "yes" ]
			then
				echo "okay, may fortune favour the brave!"
				aws --profile=${cli_profile} s3api put-bucket-lifecycle \
					--bucket ${bucket_name} --lifecycle-configuration file://$TMPDIR/s3_to_ia_lifecycle.json
			else
				echo "in the twilight zone between truth & falsehood. moving on to the next bucket ..."
				continue
			fi
			
		else
			echo "found this; leaving the bucket lifecycle as-is: $lc_cfg"
		fi
	done
}

# -----------------------------------------------------------------------------------
# defaults
cli_profile=""
region=""
days=30
lc_json_ref_fname="s3_to_ia_lifecycle.json" 
lc_json_work_fname="$TMPDIR/s3_to_ia_lifecycle.json" 

# -----------------------------------------------------------------------------------
# fetch command arguments
while getopts ":p:r:d:" o; do
	case "${o}" in
		p)
			cli_profile=${OPTARG}
			;;
		r)
			region=${OPTARG}
			;;
		d)
			days=${OPTARG}
			if [ $days -lt 30 ]
			then
				echo "too few days for the lifecycle rule"
				usage
				exit 2
			fi
			;;
		*)
			echo "unknown option ${o}"
			usage
			exit 3
			;;
	esac
done
shift $((OPTIND-1))

if [ -z "${cli_profile}" ] || [ -z "${region}" ]; then
	usage
	exit 4
fi

# echo "cli_profile = ${cli_profile}"
# echo "region = ${region}"
# echo "days = ${days}"

# -----------------------------------------------------------------------------------
# create the lifecycle JSON with the necessary transition#days, from the reference file which has a default of 30 days.

echo "creating the lifecycle JSON with $days days transition ..."
sed "s/\"Days\": 30/\"Days\": ${days}/" < $lc_json_ref_fname > $lc_json_work_fname

echo "$lc_json_work_fname:"
cat $lc_json_work_fname

echo "... done."

# -----------------------------------------------------------------------------------
# main

check_and_set_lifecycle_config

# -----------------------------------------------------------------------------------
# cleanup
rm -f $lc_json_work_fname

exit 0
