#!/bin/bash

# Notes:
# 1. Week starts with Sunday (0) and ends on Saturday (6). This is inline with Unix 'date' command.
#
# 2. A time slot cannot wrap around from Saturday to Sunday/further. It needs to be broken up at Sat 23:59:59 and further if needed.
# 3. Time slots should be sorted in time order (duh!)

# The following for Bash v4 & higher
# declare -A day_nums
# day_nums=(['Sun']=0 ['Mon']=1 ['Tue']=2 ['Wed']=3 ['Thu']=4 ['Fri']=5 ['Sat']=6)

bash_ver=$(echo $BASH_VERSION | cut -c1)

if [ $bash_ver -eq 3 ]
then
	declare "day_nums_Sun=0"
	declare "day_nums_Mon=1"
	declare "day_nums_Tue=2"
	declare "day_nums_Wed=3"
	declare "day_nums_Thu=4"
	declare "day_nums_Fri=5"
	declare "day_nums_Sat=6"
	
	# Get a value:
	arrayGet() { 
	    local array=$1 index=$2
	    local i="${array}_$index"
	    arrayGet_rv="${!i}"
	}
	
fi

time_slots=(
	[0]="Sun 00:00:00 - Mon 06:00:00" \
	[1]="Mon 06:00:01 - Mon 18:00:00" \
	[2]="Mon 18:00:01 - Tue 06:00:00" \
	[3]="Tue 06:00:01 - Tue 18:00:00" \
	[4]="Tue 18:00:01 - Wed 00:00:00" \
	[5]="Wed 00:00:01 - Wed 23:00:00" \
	[6]="Wed 23:00:01 - Thu 12:00:00" \
	[7]="Thu 12:00:01 - Fri 00:00:00" \
	[8]="Fri 00:00:01 - Sat 23:59:59"
)


# time_to_slot:
# arg 1: a date string like "+%a %H:%M:%S" i.e., "Mon 14:05:32"
# return value: a time slot from the time_slots array say ${time_slots[4]}

function time_to_slot()
{
	day_time=($1)	# say, $1 = Mon 14:05:32

	day=${day_time[0]} # Mon
	hms=${day_time[1]} # 14:05:32

	for slot_idx in "${!time_slots[@]}"
	do
		# get current slot's begin & end
		slot=${time_slots[$slot_idx]}     # say, slot = "Mon 06:00:01 - Mon 18:00:00"
		slot_beg_end=(${slot/ -/}) # drop " -" to make all fields single space separated

		slot_beg_day=${slot_beg_end[0]}   # Mon
		slot_beg_hms=${slot_beg_end[1]}   # 06:00:01

		slot_end_day=${slot_beg_end[2]}   # Mon
		slot_end_hms=${slot_beg_end[3]}   # 18:00:00

		# now see if day_time is within this slot

		if [ $bash_ver -eq 3 ]
		then
			# associative array hack for Bash v3
			arrayGet day_nums $day
			day_num=$arrayGet_rv

			arrayGet day_nums $slot_beg_day
			slot_beg_day_num=$arrayGet_rv

			arrayGet day_nums $slot_end_day
			slot_end_day_num=$arrayGet_rv

		elif [ $bash_ver -eq 4 ]
		then
			day_num=${day_nums[$day]}
			slot_beg_day_num=${day_nums[$slot_beg_day]}
			slot_end_day_num=${day_nums[$slot_end_day]}
		else
			echo "strange bash version - $BASH_VERSION; neither 3.x nor 4.x; bailing out"
			exit -1
		fi

		if [ $day_num -lt $slot_beg_day_num ] || [ $day_num -gt $slot_end_day_num ]
		then
			continue
		fi

		# if the day_num+hms is past the slot's begin, then we declare a slot match if either day_num is smaller than the slot's end day -or- if it is the same day and the hms is smaller than the slot's end hms.
		if [ $day_num -eq $slot_beg_day_num ] && [ $day_num -eq $slot_end_day_num ]
		then
			if [ $hms = $slot_beg_hms -o $hms \> $slot_beg_hms ] && [ $hms = $slot_end_hms -o $hms \< $slot_end_hms ]
			then
				# slot beg & end are on same day and
				# hms falls within that day's window
				return $slot_idx
			else
				continue # hms falls outside the matching day's window
			fi
		fi

		if [ $day_num -gt $slot_beg_day_num ] && [ $day_num -lt $slot_end_day_num ]
		then
			# day_num falls completely inside beg/end days of the window
			return $slot_idx
		fi

		if [ $day_num -eq $slot_beg_day_num ] && [ $hms = $slot_beg_hms -o $hms \> $slot_beg_hms ]
		then
			# matches beg_day, and starts after beg_day hms; so ...
			return $slot_idx
		fi

		if [ $day_num -eq $slot_end_day_num ] && [ $hms = $slot_end_hms -o $hms \< $slot_end_hms ]
		then
			# matches end_day, and ends before end_day hms; so ...
			return $slot_idx
		fi
	done

	return -1
}

echo "# Time Slots" 
for i in "${!time_slots[@]}"
do
	echo ${time_slots[$i]}
done

echo ""


# Tests ======================
echo "# Tests ======================"

# 0. Smoke
echo "# 0. Smoke"

t="Mon 18:00:01"
time_to_slot "$t"
echo "$t: slot_idx = $?; slot = ${time_slots[$?]}"

echo ""

# 1. Boundary values
echo "# 1. Boundary values"
for i in "${!time_slots[@]}"
do
	slot=${time_slots[$i]}
	slot_beg=$(echo $slot | cut -d'-' -f1 | sed 's/ $//') # strip trailing space
	slot_end=$(echo $slot | cut -d'-' -f2 | sed 's/^ //') # strip leading space

	time_to_slot "$slot_beg"
	echo "$slot_beg: slot_idx = $?; slot = ${time_slots[$?]}"

	time_to_slot "$slot_end"
	echo "$slot_end: slot_idx = $?; slot = ${time_slots[$?]}"
done

echo ""

# 2. Current time
echo "# 2. Current time"
cur_time=$(date "+%a %H:%M:%S")

time_to_slot "$cur_time"
echo "$cur_time: slot_idx = $?; slot = ${time_slots[$?]}"

echo ""

# 3. Day starts and ends
echo "# 3. Day starts and ends"
time_to_slot "Sun 00:00:00"
echo "Sun 00:00:00: slot_idx = $?; slot = ${time_slots[$?]}"
time_to_slot "Sun 23:59:59"
echo "Sun 23:59:59: slot_idx = $?; slot = ${time_slots[$?]}"

time_to_slot "Mon 00:00:00"
echo "Mon 00:00:00: slot_idx = $?; slot = ${time_slots[$?]}"
time_to_slot "Mon 23:59:59"
echo "Mon 23:59:59: slot_idx = $?; slot = ${time_slots[$?]}"

time_to_slot "Tue 00:00:00"
echo "Tue 00:00:00: slot_idx = $?; slot = ${time_slots[$?]}"
time_to_slot "Tue 23:59:59"
echo "Tue 23:59:59: slot_idx = $?; slot = ${time_slots[$?]}"

time_to_slot "Wed 00:00:00"
echo "Wed 00:00:00: slot_idx = $?; slot = ${time_slots[$?]}"
time_to_slot "Wed 23:59:59"
echo "Wed 23:59:59: slot_idx = $?; slot = ${time_slots[$?]}"

time_to_slot "Thu 00:00:00"
echo "Thu 00:00:00: slot_idx = $?; slot = ${time_slots[$?]}"
time_to_slot "Thu 23:59:59"
echo "Thu 23:59:59: slot_idx = $?; slot = ${time_slots[$?]}"

time_to_slot "Fri 00:00:00"
echo "Fri 00:00:00: slot_idx = $?; slot = ${time_slots[$?]}"
time_to_slot "Fri 23:59:59"
echo "Fri 23:59:59: slot_idx = $?; slot = ${time_slots[$?]}"

time_to_slot "Sat 00:00:00"
echo "Sat 00:00:00: slot_idx = $?; slot = ${time_slots[$?]}"
time_to_slot "Sat 23:59:59"
echo "Sat 23:59:59: slot_idx = $?; slot = ${time_slots[$?]}"

echo ""

# 3. Mid days
echo "# 3. Mid days"

time_to_slot "Sun 12:00:00"
echo "Sun 12:00:00: slot_idx = $?; slot = ${time_slots[$?]}"

time_to_slot "Mon 12:00:00"
echo "Mon 12:00:00: slot_idx = $?; slot = ${time_slots[$?]}"

time_to_slot "Tue 12:00:00"
echo "Tue 12:00:00: slot_idx = $?; slot = ${time_slots[$?]}"

time_to_slot "Wed 12:00:00"
echo "Wed 12:00:00: slot_idx = $?; slot = ${time_slots[$?]}"

time_to_slot "Thu 12:00:00"
echo "Thu 12:00:00: slot_idx = $?; slot = ${time_slots[$?]}"

time_to_slot "Fri 12:00:00"
echo "Fri 12:00:00: slot_idx = $?; slot = ${time_slots[$?]}"

time_to_slot "Sat 12:00:00"
echo "Sat 12:00:00: slot_idx = $?; slot = ${time_slots[$?]}"

echo ""
