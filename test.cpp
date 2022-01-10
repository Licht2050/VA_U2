#include <iostream>
#include "kugelkondensator.h"

using namespace std;


int main(){
    
    Kondensator kKondensator = {0.0, 0.0};

    bool pruefe = userInput (kKondensator);

    cout << kKondensator.inRadius << endl;


    return 0;
}