import json
import os
from datetime import datetime, timedelta

def generate_folders(start_date, end_date):
    folders = []
    current_date = start_date
    while current_date <= end_date:
        folder_name = current_date.strftime("%Y_%m_%d")
        folder_path = os.path.join("C:\\defects", folder_name)
        folders.append(folder_path)
        current_date += timedelta(days=1)
    return folders

def main():
    start_date = datetime(2024, 4, 21)
    end_date = datetime(2024, 6, 30)
    folders = generate_folders(start_date, end_date)
    with open("folders.json", "w") as json_file:
        json.dump(folders, json_file, indent=4)

if __name__ == "__main__":
    main()
