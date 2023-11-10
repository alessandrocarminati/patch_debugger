# Patch Debugger
## Overview

The Patch Debugger is a tool designed to help you understand why a patch 
might not be applying successfully. It provides insights into the matching 
process between the patch and the target text, helping you identify issues 
and improve the patch.

## Features
Match Scoring: The debugger evaluates the match score for each hunk in the 
patch against the target text. The match score indicates how well the hunk 
aligns with the text.

Positional Constraints: The debugger allows you to set a specific position 
in the text where the matching process begins. This helps you focus on 
particular sections of the text and understand the impact of different 
starting positions.

Penalties and Rewards: The scoring system includes penalties for missing 
lines and non-consecutive matches. It also rewards positive matches based 
on the specified position.

## Usage
1. Install Go (if not already installed)
Make sure you have Go installed on your system. 
You can download it from https://golang.org/dl/.

2. Clone the Repository
Clone the Patch Debugger repository to your local machine

3. Make and Run the Debugger
Use the following commands to run the debugger:

now it is quite easy to build:
```
go build main.go
```
run is also easy
```
./main -patch file.patch
```
4. View Results
Run the debugger against a patch results in a colored log. 
```
Processing hunk #1 on file drivers/opp/core.c
hunk #1 does NOT appily
 }
 
 /**
- * dev_pm_opp_set_clkname() - Set clk name for the device
- * @dev: Device for which clk name is being set.
- * @name: Clk name.
- *
- * In order to support OPP switching, OPP layer needs to get pointer to the
- * clock for the device. Simple cases work fine without using this routine (i.e.
- * by passing connection-id as NULL), but for a device with multiple clocks
- * available, the OPP core needs to know the exact name of the clk to use.
  *
  * This must be called before any OPPs are initialized for the device.
  */

[...]
-
-	opp_table = dev_pm_opp_set_clkname(dev, name);
-	if (IS_ERR(opp_table))
-		return PTR_ERR(opp_table);
-
-	return devm_add_action_or_reset(dev, devm_pm_opp_clkname_release,
-					opp_table);
 }
-EXPORT_SYMBOL_GPL(devm_pm_opp_set_clkname);
 
 /**
  * dev_pm_opp_register_set_opp_helper() - Register custom set OPP helper
Processing hunk #2 on file drivers/opp/core.c
hunk #2 appiles with offset 108
Processing hunk #3 on file drivers/opp/core.c
hunk #3 appiles with offset 108
```
