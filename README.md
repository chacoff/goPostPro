# Current code running without MES

# Introduction 

The goal of the software is to automatically post process the data from Dias Pyrosoft. Pyrosoft stored several plain data files containing 
information of the temperature of the sheet piles processed in Mill 2, Belval.

Our objective is to build a solution to automatically post process the data according the following restrictions: 

- the data has to be processed according the *cage* (according each camera also known as DUO)
- the data is link to a *montage* number. Because there is a config file link to each *montage*
- We query in the database to understand the start/end time of each *montage* during the day
- We sort the data according the montage and we post process it during the corresponding config file
- The post process output is stored in a share folder with csv format (comma separated)

![alt text](includes/DIAS_autoapp.drawio.png)



## The project got some modification after the Internship of Pierrick : Here are the modification :

Now the project work like this :

![alt text](includes/New_DIAS_app.png)

0 The main service application is the body of the apllication, this is from here that everything is launch  
1 Launching the file observer to get the files in live and the server for MES to connect their client and start the communication  
2 Stock the files in the file observer  
3 Receive the message from MES  
4 Send the data contain in the message to the main application  
5 When the main application recieve the data from MES, send a ping to the file observer for it to send the files that match with the MES message.  
(The MES message arrive at the end of all the passes of a cage. So the file observer will send only the files of the rolling campain just finished and restart to watch new files)  
6  Process the files and send the post processing data to the server  
7 Send the data of the post process to MES for them to put it in ArcelorMittal's data base  

file observer : A file observer is an application that continusly get the information if a file is created / modified of deleted on the computer  
Server : A server is a unique application that can handle connection from client to send data between multiple computer  



##  
## Database

| Device | Address | Catalog | Login |
| --- | --- | --- | --- |
| database | esc-sql-mi01 | MI_FDS_TR2 | EUROPE\ $user |

##  
##  Computers credentials with VNC

check with IT department to have access with VNC using Windows Login


| Device | DUO01-02 | DUO03 | DUO04 |
| --- | --- | --- | --- |
| PC |  8CC1482L9C |  8CC1482LBJ  |  8CC1482L90  |
| user |  EBT2PC21 |  EBT2PC22  |  EBT2PC25  |
| pass |  Arcelor+2021 |  Arcelor+2021  |  Arcelor+2021  |
| IP AM | 10.28.114.89 | 10.28.114.56 | 10.28.100.0 |
| IP DIAS |  16.22.1.20 |  16.22.1.100  |  16.25.1.100  |
| cam1 |  16.22.1.21 |  16.22.1.10  |  16.25.1.10  |
| cam2 |  16.22.1.22 |  |  |

## Belval Network

![alt text](includes/DIAS_Belval.drawio.png)
             