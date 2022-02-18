# Sensor


Prochot: PROCessor HOT

現在很多筆記本廠家都引入了一種名叫BD PROCHOT(Bi-directional processor hot)的功能來解決高端GPU和CPU的發熱問題。其核心原理就是在獨立GPU工作的時候,當CPU溫度超過某一閾值,就自動關閉睿頻或是降頻以達到減少發熱、節省功耗的目的。

> https://www.reddit.com/r/pcmasterrace/comments/39wmrw/throttlestop_prochot_and_you/
>
So as we all know Intel has a specified throttling temperatures for their CPUs, most of them, if not all, are designed to operate up to 100C, any higher than that at CPU starts to throttle. What Intel also does is they allow laptop manufacturers ( not sure about motherboards and desktops ) to set an offset to this throttle temperature all the way down to 85C, which in case of laptops can and will severely limit their performance. The reasoning behind that is they probably want to prolong the life of the laptop, but its actually just bullshit, laptops can operate easily at around 90-95C, my old laptop with and i5 CPU served me well for 4 years at 95C and its still working, my new laptop however ( from the same manufacturer ) has its CPU temp limit set to 85C.

PROCHOT stands for Processor Hot and is a trigger temperature at which CPU starts to throttle, you can view it by using a programm "ThrottleStop" version 7 or 8, but usually it can only be changed through BIOS. Most laptop BIOS are locked and require a modified BIOS. But one thing i found is when booting windows after entering BIOS my PROCHOT temp is set to 95 instead of 85, which results in a way better and stable performance, at a cost of a slightly hotter CPU.

The reason is that higher performance and laptop temperature mean higher power draw from the battery, severely reducing battery life, negating the whole "portable" part of a laptop. Also, since it's called a laptop for a reason, you don't want it running at 100C in your lap, or on the desk really, again because it's portable, meaning smaller heatsinks, meaning higher temperatures and power draw at higher performance.
