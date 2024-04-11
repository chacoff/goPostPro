# using Pkg
# Pkg.add("DataFrames")
# Pkg.add("TimeZones")
# Pkg.add("CompoundPeriods")
# Pkg.add("TimesDates")
# precompile
using TimesDates, Dates
using DataFrames
using Plots

function readFile(filename::AbstractString)
    
    timestamps = []
    temperatures = []
    
    file = open(filename, "r")
    readline(file)
    
    try
        while !eof(file)
            line = readline(file)  # skip the first line
            line = replace(line, "," => ".")
            elements = split(line)
            
            if length(elements) >= 2    
                times = join(elements[1:2], "T")
                timeStamp = TimeDate(times)

                temps = join(elements[8:end-4], " ")  # end-4
                tempsArray = [parse(Float64, substr) for substr in split(temps)]

                push!(timestamps, timeStamp)
                push!(temperatures, tempsArray)
            end
        end
    finally
        close(file)
    end

    return timestamps, temperatures
end

function extractData(start_timestamp::TimeDate, end_timestamp::TimeDate)

    dff = filter(row -> row.Times >= start_timestamp && row.Times <= end_timestamp, df)
<<<<<<< HEAD
=======
    # dff_x = dff.Temps    
    # heatmap(hcat(dff_x...))
>>>>>>> 1c0433198cf0a53d9a63141cd441499ca84bdad1

    return dff
end


## entry point:
filename = "C:\\Users\\gomezja\\OneDrive - ArcelorMittal\\Documents\\00_Dev\\GoPostPro\\juliaPostPro\\DUO01-02_0891_half.txt"
stamps, temps = readFile(filename)

df = DataFrame(Times=stamps, Temps=temps)
<<<<<<< HEAD
df_selection = extractData(TimeDate("2024-02-13T11:05:06.600"), TimeDate("2024-02-13T11:05:08.957"))
print(df_selection)
dff_x = df_selection.Temps
heatmap(hcat(dff_x...), title="heatmap")
=======
println(df)
dff = extractData(TimeDate("2024-02-13T11:05:06.100"), TimeDate("2024-02-13T11:05:16.000"))
println(dff)
println("Ok!")
>>>>>>> 1c0433198cf0a53d9a63141cd441499ca84bdad1


