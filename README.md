# tomcalendar

Parses a text file of date-based reminders and prints those that fall within a given date range.

I used to use OpenBSD's calendar(1), but I kept missing reminders because that program silently ignores most errors. tomcalendar prints error messages in the cases that trippd me up, and also supports some new features for specifying dates.

## File format

One line per reminder, consisting of a date followed by the reminder text, separated by a tab character, like "January 1, 2021	A strange New Year's". There's a lot you can do, but every feature is a variation on that theme: [date] [tab] [reminder].

Here's an example file that illustrates all the features:

    January 1, 2021	Occurs one specific day
    	Second reminder for that day (a blank date takes the previous line's date)
    December 25
    	Occurs on Christmas of every year (the date can go on a line by itself)
    	
    *	Occurs every day
    */3	Occurs every third day (starting January 1, 1970)
    Sunday	Occurs every Sunday
    Friday/2	Occurs every other Friday
    Sunday | Friday	Occurs every Sunday and Friday
    
    1 *	Occurs on the first day of every month
    -2 *	Occurs on the second-to-last day of every month (e.g. January 30)

## Usage

    tomcalendar < calendarfile
    tomcalendar -calendar calendarfile

Prints reminders for today's date using the contents of calendarfile.

    tomcalendar -calendar calendarfile -date 2020-03-22

Prints the reminders for a specific date

    tomcalendar -calendar calendarfile -since 2020-03-22

Prints reminders for every date between 2020-03-22 and the current date, not including 2020-03-22, and printing every reminder at most once.

-since is useful if you use tomcalendar ~daily but forget to run it some days. For example, my `agenda` script (written in [rc](https://en.wikipedia.org/wiki/Rc), not bash) looks like this:

    #!/usr/local/bin/rc
    today=`{/bin/date '+%Y-%m-%d'}
    echo './agenda -since' $today
    tomcalendar $* -calendar calendarfile

So every time I run it, it prints the command I should run next time to get the reminders since the last time I ran it.
