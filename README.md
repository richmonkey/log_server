Introduction
---------
A line log server, every line log is appended to a file.
The single log file is limited in 1GB. if tag's directory name not exist, it will auto create.

run server `./log_server log.cfg -log_dir="./log" -stderrthreshold=ERROR`
