import os
import time
import multiprocessing

def process_task(process_id, num_processes, filename, token_queue, shutdown_event):
    """Function executed by each process to write to a shared file using a token-ring mechanism."""
    while not shutdown_event.is_set():
        token = token_queue.get()  # Wait for token

        if token == "STOP":
            break  # Exit process cleanly

        # Critical Section: Write to the file
        with open(filename, "a") as f:
            f.write(f"Process {process_id} writing at {time.strftime('%X')}\n")
            print(f"Process {process_id} wrote to the file.")

        time.sleep(1)  # Simulate some work

        # Pass the token to the next process in the ring
        next_process = (process_id + 1) % num_processes
        token_queue.put("TOKEN")

def main():
    num_processes = 5  # Number of processes in the ring
    filename = "shared_file.txt"

    # Ensure the file is empty at the start
    open(filename, "w").close()

    # Create a Queue for passing the token
    token_queue = multiprocessing.Queue()
    
    # Shutdown event for clean exit
    shutdown_event = multiprocessing.Event()

    # Create and start processes
    processes = []
    for i in range(num_processes):
        p = multiprocessing.Process(target=process_task, args=(i, num_processes, filename, token_queue, shutdown_event))
        processes.append(p)
        p.start()

    # Send the initial token to the first process
    token_queue.put("TOKEN")

    # Let processes run for some time
    time.sleep(10)

    # Signal shutdown
    shutdown_event.set()

    # Send stop signal for each process
    for _ in range(num_processes):
        token_queue.put("STOP")

    # Wait for processes to finish
    for p in processes:
        p.join()

    print("Processes terminated. Check the shared file for output.")

if __name__ == "__main__":
    main()
