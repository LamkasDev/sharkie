import idaapi
import idc
import idautils
import os

COLOR_COVERED_ASM  = 0xCCFFCC  # Light Green
COLOR_COVERED_FUNC = 0xCCFFCC  # Light Green
COLOR_CALLER       = 0xDDBB99  # Light Blue

def get_coverage_path():
    """
    Constructs the coverage filename based on the current IDB name
    and the script's location.
    """
    try:
        script_dir = os.path.dirname(os.path.abspath(__file__))
    except NameError:
        print("[!] Error: __file__ undefined. Please run this script via 'File -> Script File'.")
        return None

    root_name = idaapi.get_root_filename()
    return os.path.join(script_dir, root_name + ".coverage")

def color_callers(func_start):
    """
    Finds all code references TO this function and colors them.
    """
    # CodeRefsTo(ea, 1) returns all addresses that branch/call to 'ea'
    for caller_ea in idautils.CodeRefsTo(func_start, 1):
        # Only color it if it hasn't been colored as 'executed' yet.
        # We prioritize "Green" (Executed) over "Blue" (Caller).
        current_color = idc.get_color(caller_ea, idc.CIC_ITEM)
        if current_color != COLOR_COVERED_ASM:
            idc.set_color(caller_ea, idc.CIC_ITEM, COLOR_CALLER)

def main():
    cov_file = get_coverage_path()
    if not cov_file:
        return

    print(f"[*] Attempting to load coverage from {cov_file}")
    if not os.path.exists(cov_file):
        print(f"[!] Error: Coverage file {cov_file} not found")
        return

    count_asm = 0
    count_funcs = 0
    processed_funcs = set()
    with open(cov_file, "r") as f:
        lines = f.readlines()

    print(f"[*] Processing {len(lines)} trace entries...")
    for line in lines:
        line = line.strip()
        if not line: continue
        try:
            # Handle both "0x1234" and "1234" formats
            ea = int(line, 16)

            # Color the executed instruction (Green)
            idc.set_color(ea, idc.CIC_ITEM, COLOR_COVERED_ASM)
            count_asm += 1

            # Identify the function this instruction belongs to
            func_start = idc.get_func_attr(ea, idc.FUNCATTR_START)

            if func_start != idc.BADADDR and func_start not in processed_funcs:
                # Color the function entry in the sidebar/graph view
                idc.set_color(func_start, idc.CIC_FUNC, COLOR_COVERED_FUNC)

                # Color the Callers (Blue)
                color_callers(func_start)

                processed_funcs.add(func_start)
                count_funcs += 1

        except ValueError:
            print(f"[!] Invalid hex line: {line}")
        except Exception as e:
            print(f"[!] Error processing {line}: {e}")

    print("-" * 40)
    print(f"Done! Colored {count_asm} instructions.")
    print(f"Covered {count_funcs} unique functions (and highlighted their callers).")
    idaapi.refresh_idaview_anyway()

if __name__ == "__main__":
    main()