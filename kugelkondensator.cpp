#include <iostream>
#include "kugelkondensator.h"

using namespace std;

bool userInput (Kondensator & k){
    cout << "InnerenRadius-Aussenradius: ";
    cin >> k.inRadius >> k.ausRadius;
}

double berechnung(const Kondensator & k){
    return 4 * PI * EPSILON_NULL 
				* k.ausRadius * k.inRadius / (k.ausRadius - k.inRadius);
}